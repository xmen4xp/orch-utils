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
	// OnDeletePolicy is the value of the nexus-on-delete tag on the parent
	// field that declares this node as a child (e.g. "restrict"). Empty means
	// the default, "cascade".
	OnDeletePolicy string
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
	// OnDeletePolicy mirrors Node.OnDeletePolicy for this child edge. It is
	// compiler-internal (json:"-"): the api-gw enforces via the precomputed
	// deletion-restrict-children list on the CRD, not per-child policy, so this
	// is deliberately kept out of the serialized CRD annotation.
	OnDeletePolicy string `json:"-"`
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
