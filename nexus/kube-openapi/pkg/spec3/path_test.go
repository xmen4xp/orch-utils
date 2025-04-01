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

func TestPathJSONSerialization(t *testing.T) {
	cases := []struct {
		name           string
		target         *spec3.Path
		expectedOutput string
	}{
		{
			name: "basic",
			target: &spec3.Path{
				PathProps: spec3.PathProps{
					Get: &spec3.Operation{
						OperationProps: spec3.OperationProps{
							Description: "Returns pets based on ID",
							Summary:     "Find pets by ID",
							OperationId: "getPetsById",
							Responses: &spec3.Responses{
								ResponsesProps: spec3.ResponsesProps{
									StatusCodeResponses: map[int]*spec3.Response{
										200: &spec3.Response{
											ResponseProps: spec3.ResponseProps{
												Description: "Pet response",
												Content: map[string]*spec3.MediaType{
													"*/*": &spec3.MediaType{
														MediaTypeProps: spec3.MediaTypeProps{
															Schema: &spec.Schema{
																SchemaProps: spec.SchemaProps{
																	Type: []string{"array"},
																	Items: &spec.SchemaOrArray{
																		Schema: &spec.Schema{
																			SchemaProps: spec.SchemaProps{
																				Ref: spec.MustCreateRef("#/components/schemas/Pet"),
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Parameters: []*spec3.Parameter{
						&spec3.Parameter{
							ParameterProps: spec3.ParameterProps{
								Name:        "id",
								In:          "path",
								Description: "ID of the pet to use",
								Required:    true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"array"},
										Items: &spec.SchemaOrArray{
											Schema: &spec.Schema{
												SchemaProps: spec.SchemaProps{
													Ref: spec.MustCreateRef("#/components/schemas/Pet"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedOutput: `{"get":{"summary":"Find pets by ID","description":"Returns pets based on ID","operationId":"getPetsById","responses":{"200":{"description":"Pet response","content":{"*/*":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/Pet"}}}}}}},"parameters":[{"name":"id","in":"path","description":"ID of the pet to use","required":true,"schema":{"type":"array","items":{"$ref":"#/components/schemas/Pet"}}}]}`,
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
				t.Fatalf("%s", serializedTarget)
				t.Fatalf("diff %s", cmp.Diff(serializedTarget, tc.expectedOutput))
			}
		})
	}
}
