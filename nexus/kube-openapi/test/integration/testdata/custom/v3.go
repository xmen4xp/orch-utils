// Copyright 2019 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package custom

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/common"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
)

// +k8s:openapi-gen=true
type Bal struct{}

func (_ Bal) OpenAPIV3Definition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		},
	}
}

// +k8s:openapi-gen=true
type Bac struct{}

func (_ Bac) OpenAPIV3Definition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
			},
		},
	}
}

func (_ Bac) OpenAPIDefinition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		},
	}
}

// +k8s:openapi-gen=true
type Bah struct{}

func (_ Bah) OpenAPIV3Definition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
			},
		},
	}
}

func (_ Bah) OpenAPISchemaType() []string {
	return []string{"test-type"}
}

func (_ Bah) OpenAPISchemaFormat() string {
	return "test-format"
}

// FooV3OneOf has an OpenAPIV3OneOfTypes method
// +k8s:openapi-gen=true
type FooV3OneOf struct{}

func (FooV3OneOf) OpenAPIV3OneOfTypes() []string {
	return []string{"number", "string"}
}
func (FooV3OneOf) OpenAPISchemaType() []string {
	return []string{"string"}
}
func (FooV3OneOf) OpenAPISchemaFormat() string {
	return "string"
}
