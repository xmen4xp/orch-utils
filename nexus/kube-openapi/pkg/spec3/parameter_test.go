// Copyright 2021 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec3_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/spec3"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
)

func TestParameterJSONSerialization(t *testing.T) {
	cases := []struct {
		name           string
		target         *spec3.Parameter
		expectedOutput string
	}{
		{
			name: "header parameter",
			target: &spec3.Parameter{
				ParameterProps: spec3.ParameterProps{
					Name:        "token",
					In:          "header",
					Description: "token to be passed as a header",
					Required:    true,
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
					Style: "simple",
				},
			},
			expectedOutput: `{"name":"token","in":"header","description":"token to be passed as a header","required":true,"style":"simple","schema":{"type":"integer","format":"int64"}}`,
		},
		{
			name: "path parameter",
			target: &spec3.Parameter{
				ParameterProps: spec3.ParameterProps{
					Name:        "username",
					In:          "path",
					Description: "username to fetch",
					Required:    true,
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"string"},
						},
					},
				},
			},
			expectedOutput: `{"name":"username","in":"path","description":"username to fetch","required":true,"schema":{"type":"string"}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rawTarget, err := json.Marshal(tc.target)
			if err != nil {
				t.Fatal(err)
			}
			serializedTarget := string(rawTarget)
			if !cmp.Equal(serializedTarget, tc.expectedOutput) {
				t.Fatalf("diff %s", cmp.Diff(serializedTarget, tc.expectedOutput))
			}
		})
	}
}
