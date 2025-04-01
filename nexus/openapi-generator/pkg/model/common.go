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
	IgnoredParentPathParams []string `yaml:"ignoredParentPathParams"`
}

var OpenApiIgnoredParentPathParams map[string]struct{} = make(map[string]struct{})

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

	for _, param := range config.IgnoredParentPathParams {
		OpenApiIgnoredParentPathParams[param] = struct{}{}
		fmt.Println("adding ignored param :", param)
	}
}
