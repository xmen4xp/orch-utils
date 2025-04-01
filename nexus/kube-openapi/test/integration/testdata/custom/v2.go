// Copyright 2019 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package custom

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/common"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
)

// +k8s:openapi-gen=true
type Bak struct{}

func (_ Bak) OpenAPIDefinition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"integer"},
			},
		},
	}
}
