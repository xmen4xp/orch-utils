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

func TestServerJSONSerialization(t *testing.T) {
	cases := []struct {
		name           string
		target         *spec3.Server
		expectedOutput string
	}{
		// scenario 1
		{
			name: "basic",
			target: &spec3.Server{
				ServerProps: spec3.ServerProps{
					URL:         "https://development.gigantic-server.com/v1",
					Description: "Development server",
				},
			},
			expectedOutput: `{"description":"Development server","url":"https://development.gigantic-server.com/v1"}`,
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
