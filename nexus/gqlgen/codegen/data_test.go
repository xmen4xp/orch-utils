// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/codegen/config"

	"github.com/stretchr/testify/assert"
)

func TestData_Directives(t *testing.T) {
	d := Data{
		Config: &config.Config{
			Sources: []*ast.Source{
				{
					Name: "schema.graphql",
				},
			},
		},
		AllDirectives: DirectiveList{
			"includeDirective": {
				DirectiveDefinition: &ast.DirectiveDefinition{
					Name: "includeDirective",
					Position: &ast.Position{
						Src: &ast.Source{
							Name: "schema.graphql",
						},
					},
				},
				Name:    "includeDirective",
				Args:    nil,
				Builtin: false,
			},
			"excludeDirective": {
				DirectiveDefinition: &ast.DirectiveDefinition{
					Name: "excludeDirective",
					Position: &ast.Position{
						Src: &ast.Source{
							Name: "anothersource.graphql",
						},
					},
				},
				Name:    "excludeDirective",
				Args:    nil,
				Builtin: false,
			},
		},
	}

	expected := DirectiveList{
		"includeDirective": {
			DirectiveDefinition: &ast.DirectiveDefinition{
				Name: "includeDirective",
				Position: &ast.Position{
					Src: &ast.Source{
						Name: "schema.graphql",
					},
				},
			},
			Name:    "includeDirective",
			Args:    nil,
			Builtin: false,
		},
	}

	assert.Equal(t, expected, d.Directives())
}
