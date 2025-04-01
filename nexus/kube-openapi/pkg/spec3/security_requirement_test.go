// Copyright 2021 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec3_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/spec3"
)

func TestSecurityRequirementJSONSerialization(t *testing.T) {
	cases := []struct {
		name           string
		target         *spec3.SecurityRequirement
		expectedOutput string
	}{
		{
			name: "Non-OAuth2 Security Requirement",
			target: &spec3.SecurityRequirement{
				SecurityRequirementProps: map[string][]string{
					"api_key": []string{},
				},
			},
			expectedOutput: `{"api_key":[]}`,
		},
		{
			name: "OAuth2 Security Requirement",
			target: &spec3.SecurityRequirement{
				SecurityRequirementProps: map[string][]string{
					"petstore_auth": []string{
						"write_pets",
						"read:pets",
					},
				},
			},
			expectedOutput: `{"petstore_auth":["write_pets","read:pets"]}`,
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
