// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"go/ast"

	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Node struct {
	Name             string
	PkgName          string
	FullName         string
	CrdName          string
	IsSingleton      bool
	Imports          []*ast.ImportSpec
	TypeSpec         *ast.TypeSpec
	Parents          []string
	SingleChildren   map[string]Node
	MultipleChildren map[string]Node
	SingleLink       map[string]Node
	MultipleLink     map[string]Node
	GraphqlQuerySpec nexus.GraphQLQuerySpec
	GraphqlSpec      nexus.GraphQLSpec

	// Versioning (k8s-style multi-version CRD support, see node_version.go).
	Kind                  string // Resolved Kind, shared by every version of a resource. Defaults to Name.
	Version               string // API version this struct represents (e.g. "v1", "v2"). Defaults to "v1".
	IsStorageVersion      bool   // Whether this version is the CRD's storage version.
	ConversionImplemented bool   // Whether this struct implements ConvertTo/ConvertFrom to the storage version.
	OtherVersions         []Node // Non-canonical versions of the same Kind. Only ever populated on the storage-version Node.
}

type NodeHelper struct {
	Name             string
	RestName         string
	Parents          []string
	Children         map[string]NodeHelperChild // CRD Name => NodeHelperChild
	Links            map[string]NodeHelperChild // FieldName => NodeHelperChild
	RestMappings     map[string]string
	IsSingleton      bool
	GraphqlQuerySpec nexus.GraphQLQuerySpec
	GraphqlSpec      nexus.GraphQLSpec
}

type NodeHelperChild struct {
	FieldName      string `json:"fieldName"`
	FieldNameGvk   string `json:"fieldNameGvk"`
	GoFieldNameGvk string `json:"goFieldNameGvk"`
	IsNamed        bool   `json:"isNamed"`
}

type NonNexusTypes struct {
	Types         map[string]ast.Decl
	Values        []string
	ExternalTypes []string
}

func (node *Node) Walk(fn func(node *Node)) {
	fn(node)

	for _, n := range node.MultipleChildren {
		n.Walk(fn)
	}

	for _, n := range node.SingleChildren {
		n.Walk(fn)
	}
}
