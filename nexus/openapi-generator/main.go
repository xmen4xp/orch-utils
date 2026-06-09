// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	yamlv1 "github.com/ghodss/yaml"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"nexus/openapi-generator/pkg/api"
	"nexus/openapi-generator/pkg/model"
)

func main() {
	var (
		yamlsPath           string
		extensionsPath      string
		datamodelConfigPath string
		datamodelName       string
		outputFilePath      string
	)
	flag.StringVar(&yamlsPath, "yamls-path", "", "directory containing CRD YAML definitions")
	flag.StringVar(&extensionsPath, "extensions-path", "", "directory containing ExtensionRestAPI YAML definitions (optional; defaults to <yamls-path>/../extensionrestapi)")
	flag.StringVar(&datamodelConfigPath, "datamodel-path", "", "datamodel config file")
	flag.StringVar(&datamodelName, "datamodel-name", "", "name of the datamodel")
	flag.StringVar(&outputFilePath, "output-file-path", "", "output file")

	flag.Parse()

	if yamlsPath == "" {
		panic("yamls-path is mandatory. Run with -h for help")
	}

	if datamodelName == "" {
		panic("datamodel-name is mandatory. Run with -h for help")
	}

	if datamodelConfigPath != "" {
		// Initialize configurations parameters if provided.
		model.InitOpenApiIgnoredParentPathParams(datamodelConfigPath)
	}

	// Default the extensions path to a sibling directory of the CRDs path.
	if extensionsPath == "" {
		extensionsPath = filepath.Join(filepath.Dir(yamlsPath), "extensionrestapi")
	}

	// Run openapi spec generation.
	run(yamlsPath, extensionsPath, datamodelName, outputFilePath)
}

// Generate openapi spec from nexus generated CRD yaml spec.
func run(dir, extensionsDir, datamodelName, outputFilePath string) {

	// Read all files from the input directory.
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through each of the CRD's and construct internal state.
	var nexusUris []nexus.RestURIs
	for _, file := range files {

		// Only process files with .yaml extension.
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		fileSpec, err := os.ReadFile(dir + "/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}

		crdJson, err := yamlv1.YAMLToJSON(fileSpec)
		if err != nil {
			log.Fatalf("unable to render crd spec from file %s\n", file.Name())
		}

		// Marshal file content to a CRD spec.
		var crd apiextensionsv1.CustomResourceDefinition
		err = json.Unmarshal(crdJson, &crd)
		if err != nil {
			log.Fatalf("unable to unmarshal crd spec from file %s\n", file.Name())
		}

		if nexus_ann, ok := crd.Annotations["nexus"]; ok {
			var nexus_annotation model.NexusAnnotation
			err := json.Unmarshal([]byte(nexus_ann), &nexus_annotation)
			if err != nil {
				log.Fatalf("error processing nexus annotation [%s] from file %s\n", nexus_ann, file.Name())
			}

			eventType := model.Upsert
			children := make(map[string]model.NodeHelperChild)
			if nexus_annotation.Children != nil {
				children = nexus_annotation.Children
			}

			links := make(map[string]model.NodeHelperChild)
			if nexus_annotation.Links != nil {
				links = nexus_annotation.Links
			}

			urisMap := make(map[string]model.RestURIInfo)

			// Add child, link and status URIs for each GET method
			var newUris []nexus.RestURIs
			api.ConstructNewURIs(nexus_annotation, urisMap, &newUris)
			nexus_annotation.NexusRestAPIGen.Uris = append(nexus_annotation.NexusRestAPIGen.Uris, newUris...)

			// Construct state need for openapi schema generation.
			model.ConstructMapUriToUriInfo(eventType, urisMap)
			model.ConstructMapURIToCRDType(eventType, crd.Name, nexus_annotation.NexusRestAPIGen.Uris)
			model.ConstructMapCRDTypeToNode(eventType, crd.Name, nexus_annotation.Name, nexus_annotation.Hierarchy, children, links, nexus_annotation.IsSingleton, nexus_annotation.Description, nexus_annotation.DeferredDelete)
			model.ConstructMapCRDTypeToRestUris(eventType, crd.Name, nexus_annotation.NexusRestAPIGen)
			model.ConstructMapCRDTypeToSpec(model.Upsert, crd.Name, crd.Spec)

			nexusUris = append(nexusUris, nexus_annotation.NexusRestAPIGen.Uris...)

			// Add patch API for PUT and Status methods.
			for _, v := range nexus_annotation.NexusRestAPIGen.Uris {
				if httpCodesResponse, ok := v.Methods[http.MethodPut]; ok {
					v.Methods[http.MethodPatch] = httpCodesResponse
				}
			}
		}
	}

	// Construct the openapi schema.
	api.New(datamodelName)
	for _, uri := range nexusUris {
		api.AddPath(uri, datamodelName)
	}

	// Merge ExtensionRestAPI manifests (custom endpoints like /metrics,
	// /recommendations, /cani) so that the generated spec matches what
	// api-gw serves at runtime. Missing directory is non-fatal — older
	// datamodels without extensions continue to work.
	if extensionsDir != "" {
		extFiles, err := os.ReadDir(extensionsDir)
		if err == nil {
			for _, ef := range extFiles {
				if filepath.Ext(ef.Name()) != ".yaml" {
					continue
				}
				body, err := os.ReadFile(filepath.Join(extensionsDir, ef.Name()))
				if err != nil {
					log.Printf("extension: unable to read %s: %v", ef.Name(), err)
					continue
				}
				api.AddExtensionPath(body, datamodelName)
			}
		} else if !os.IsNotExist(err) {
			log.Printf("extension: unable to read directory %s: %v", extensionsDir, err)
		}
	}

	spec, err := json.MarshalIndent(api.Schemas[datamodelName], "", "  ")
	if err != nil {
		log.Fatalf("failed to construct api schema for datamodel %s with error %v", datamodelName, err)
	}

	// If output file is specified, write the generated openapi spec to the file.
	if outputFilePath != "" {
		fmt.Println("Writing openapi spec for file ", outputFilePath)
		err := os.WriteFile(outputFilePath, spec, 0644)
		if err != nil {
			log.Fatalf("write of openapi spec to file %s failed with error %v", outputFilePath, err)
		}
	} else {
		// If not output file is specified, just return.
		fmt.Println(string(spec))
	}
}
