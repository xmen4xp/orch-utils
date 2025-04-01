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

func TestEncodingJSONSerialization(t *testing.T) {
	cases := []struct {
		name           string
		target         *spec3.Encoding
		expectedOutput string
	}{
		// scenario 1
		{
			name: "basic",
			target: &spec3.Encoding{
				EncodingProps: spec3.EncodingProps{
					ContentType: "image/png",
					Headers: map[string]*spec3.Header{
						"X-Rate-Limit-Limit": &spec3.Header{
							HeaderProps: spec3.HeaderProps{
								Description: "The number of allowed requests in the current period",
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"integer"},
									},
								},
							},
						},
					},
				},
			},
			expectedOutput: `{"contentType":"image/png","headers":{"X-Rate-Limit-Limit":{"description":"The number of allowed requests in the current period","schema":{"type":"integer"}}}}`,
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
