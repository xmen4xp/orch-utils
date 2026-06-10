// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"strings"

	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser/rest"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/util"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/config"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/generator"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/mermaid"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
)

func main() {
	configFile := flag.String("config-file", "", "Config file location.")
	dslDir := flag.String("dsl", "datamodel", "DSL file location.")
	crdDir := flag.String("crd-output", "_generated", "CRD file location.")
	logLevel := flag.String("log-level", "ERROR", "Log level")
	mermaidDir := flag.String("mermaid-output", "_generated", "Mermaid file location.")
	flag.Parse()

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Failed to configure logging: %v\n", err)
	}
	log.SetLevel(lvl)

	conf := &config.Config{}
	if *configFile != "" {
		conf, err = config.LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
	}

	if conf.CrdModulePath == "" {
		conf.CrdModulePath = "nexustempmodule/"
	}
	// env overwrites config
	envGroup := os.Getenv("GROUP_NAME")
	if envGroup != "" {
		conf.GroupName = envGroup
	}
	if conf.GroupName == "" {
		log.Fatalf("failed to determine CRD group name, please add to config file as" +
			" groupName or as GROUP_NAME enviroment variable")
	}

	config.ConfigInstance = conf
	pkgs := parser.ParseDSLPkg(*dslDir)
	graphlqQueries := parser.ParseGraphqlQuerySpecs(pkgs)
	graph, nonNexusTypes, fileset := parser.ParseDSLNodes(*dslDir, conf.GroupName, pkgs, graphlqQueries)
	methods, codes := rest.ParseResponses(pkgs)
	graphqlFiles := parser.ParseGraphQLFiles(*dslDir)

	// Create parentsMap for hierarchy lookups
	parentsMap := parser.CreateParentsMap(graph)

	// Pre-pass: parse all RestAPISpec values across all packages so their
	// PathParams aliases are published to the parser-wide alias registry
	// before ExtensionRestAPI path-param validation runs. ExtensionRestAPI
	// URIs have no PathParams map of their own and rely on aliases declared
	// by RestURIs to resolve non-formula tokens (e.g. {datacenter} ->
	// datacenters.DataCenters).
	for _, pkg := range pkgs {
		_ = rest.GetRestApiSpecs(pkg, methods, codes, parentsMap)
	}

	// Parse ExtensionRestAPI variables (basic validation happens during parsing)
	extensionRestAPIs := parser.ParseExtensionRestAPIs(pkgs)

	// Associate ExtensionRestAPI specs with nodes and validate path params
	extensionRestAPIs = associateExtensionRestAPIsWithNodes(pkgs, extensionRestAPIs, parentsMap, conf.GroupName)

	if err = generator.RenderCRDTemplate(conf.GroupName, conf.CrdModulePath, pkgs, graph,
		*crdDir, methods, codes, nonNexusTypes, fileset, graphqlFiles); err != nil {
		log.Fatalf("Error rendering crd template: %v", err)
	}

	// Render ExtensionRestAPI CR instances
	if err = generator.RenderExtensionRestAPIs(*crdDir, extensionRestAPIs, parentsMap); err != nil {
		log.Fatalf("Error rendering ExtensionRestAPI CRs: %v", err)
	}

	// Generate mermaid graph visualization
	log.Debugf("Generating mermaid graph visualization with %d root nodes to directory: %s", len(graph), *mermaidDir)
	if err = mermaid.GenerateMermaidFile(graph, *mermaidDir); err != nil {
		log.Warnf("Failed to generate mermaid graph: %v", err)
	} else {
		log.Debugf("Successfully generated mermaid graph visualization")
	}
}

// associateExtensionRestAPIsWithNodes finds the node associated with each ExtensionRestAPI
// via the // nexus-extension-rest-api:VarName annotation and validates path params.
func associateExtensionRestAPIsWithNodes(pkgs parser.Packages, specs []parser.ExtensionRestAPISpec, parentsMap map[string]parser.NodeHelper, baseGroupName string) []parser.ExtensionRestAPISpec {
	// Build a map of variable name -> spec index for quick lookup
	specMap := make(map[string]int)
	for i, spec := range specs {
		key := spec.PkgName + "." + spec.Name
		specMap[key] = i
	}

	// Scan all packages for nexus-extension-rest-api annotations on nodes
	for _, pkg := range pkgs {
		for _, node := range pkg.GetNexusNodes() {
			typeName := parser.GetTypeName(node)
			annotations, ok := parser.GetNexusExtensionRestAPIAnnotations(pkg, typeName)
			if !ok {
				continue
			}

			// Process each annotation (supports comma-separated list)
			for _, annotation := range annotations {
				// Find the corresponding ExtensionRestAPI spec
				key := pkg.Name + "." + annotation
				specIdx, found := specMap[key]
				if !found {
					log.Fatalf("ExtensionRestAPI annotation on node '%s.%s' references unknown variable '%s'",
						pkg.Name, typeName, annotation)
				}

				// Set the associated node info
				specs[specIdx].AssociatedNode = pkg.Name + "." + typeName

				// Build the CRD name for this node
				plural := strings.ToLower(util.ToPlural(typeName))
				groupName := pkg.Name + "." + baseGroupName
				specs[specIdx].NodeCRDName = plural + "." + groupName

				log.Debugf("Associated ExtensionRestAPI '%s' with node '%s' (CRD: %s)",
					annotation, specs[specIdx].AssociatedNode, specs[specIdx].NodeCRDName)
			}
		}
	}

	// Validate path params for all specs that have an associated node
	for _, spec := range specs {
		if spec.AssociatedNode != "" {
			if err := parser.ValidateExtensionRestAPIPathParams(spec, parentsMap); err != nil {
				log.Fatalf("ExtensionRestAPI '%s': %v", spec.Name, err)
			}
		}
	}

	return specs
}
