// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type EventType string

const (
	Upsert EventType = "Upsert"
	Delete EventType = "Delete"
)

type DatamodelConfig struct {
	// Title is the optional spec.info.title applied to the generated
	// OpenAPI document. When unset, the openapi-builder's built-in
	// default ("Nexus API GW APIs") is used. This field is the
	// build-time half of the single-source-of-truth contract with
	// the datamodel installer, which reads the same field at runtime
	// to populate the Datamodel CR's spec.title.
	Title                   string            `yaml:"title"`
	IgnoredParentPathParams []string          `yaml:"ignoredParentPathParams"`
	NodeToHeaderMapping     map[string]string `yaml:"nodeToHeaderMapping"`
}

// OpenApiTitle is the spec.info.title read from the datamodel
// config file (nexus.yaml). Empty means the caller should fall back
// to the openapi-builder's built-in default.
var OpenApiTitle string

var OpenApiIgnoredParentPathParams map[string]struct{} = make(map[string]struct{})

// OpenApiNodeToHeaderMapping maps a parent node type (e.g. "orgs.Org") to the
// HTTP header name (e.g. "x-org-id") that should represent it in the generated
// OpenAPI spec. Used in conjunction with OpenApiIgnoredParentPathParams: when a
// parent is ignored AND has an entry here, it is emitted as a required header
// parameter instead of being dropped.
var OpenApiNodeToHeaderMapping map[string]string = make(map[string]string)

type NexusAnnotation struct {
	Name                 string                     `json:"name,omitempty"`
	Hierarchy            []string                   `json:"hierarchy,omitempty"`
	Children             map[string]NodeHelperChild `json:"children,omitempty"`
	Links                map[string]NodeHelperChild `json:"links,omitempty"`
	NexusRestAPIGen      nexus.RestAPISpec          `json:"nexus-rest-api-gen,omitempty"`
	NexusRestAPIMappings map[string]string          `json:"nexus-rest-api-mappings,omitempty"`
	IsSingleton          bool                       `json:"is_singleton,omitempty"`
	Description          string                     `json:"description,omitempty"`
	DeferredDelete       bool                       `json:"deferred-delete,omitempty"`
}

type NodeHelperChild struct {
	FieldName    string `json:"fieldName"`
	FieldNameGvk string `json:"fieldNameGvk"`
	IsNamed      bool   `json:"isNamed"`
}

type NodeInfo struct {
	Name            string
	ParentHierarchy []string
	Children        map[string]NodeHelperChild
	Links           map[string]NodeHelperChild
	IsSingleton     bool
	Description     string
	DeferredDelete  bool
}

type RestURIInfo struct {
	TypeOfURI URIType
}

type URIType int

const (
	DefaultURI URIType = iota
	SingleLinkURI
	NamedLinkURI
	StatusURI
)

type DatamodelInfo struct {
	Title string
}

func InitOpenApiIgnoredParentPathParams(configFile string) {
	var config DatamodelConfig
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("failed to open config file %s with error %s", configFile, err)
	}
	configStr, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read config file %s with error %s", configFile, err)
	}

	err = yaml.Unmarshal(configStr, &config)
	if err != nil {
		log.Fatalf("failed to unmarshal config file %s with error %s", configFile, err)
	}

	if config.Title != "" {
		OpenApiTitle = config.Title
		fmt.Println("adding datamodel title :", config.Title)
	}

	for _, param := range config.IgnoredParentPathParams {
		OpenApiIgnoredParentPathParams[param] = struct{}{}
		fmt.Println("adding ignored param :", param)
	}

	for nodeType, headerName := range config.NodeToHeaderMapping {
		OpenApiNodeToHeaderMapping[nodeType] = headerName
		fmt.Printf("adding node->header mapping : %s -> %s\n", nodeType, headerName)
		if _, ignored := OpenApiIgnoredParentPathParams[nodeType]; !ignored {
			fmt.Printf("WARNING: nodeToHeaderMapping entry %q is not present in ignoredParentPathParams; mapping will have no effect\n", nodeType)
		}
	}
}
