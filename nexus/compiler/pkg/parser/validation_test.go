// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	"testing"

	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
)

func TestPathParamAliasFor(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"orgs.Org", "org"},
		{"projects.Project", "project"},
		{"spaces.Space", "space"},
		{"aislice.AISlice", "aislice"},
		{"credentials.K8sSecret", "k8ssecret"},
		// Single-segment input (no dot) — falls back to lowercase.
		{"Org", "org"},
		// Empty.
		{"", ""},
	}
	for _, c := range cases {
		got := parser.PathParamAliasFor(c.in)
		if got != c.want {
			t.Errorf("PathParamAliasFor(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestValidateRequiredParents_AcceptsAliasTokens(t *testing.T) {
	// Hierarchy:
	//   roots.root        (singleton, ignored)
	//   orgs.org          (parent of projects)
	//   projects.project  (the node under test — only relevant for indexing,
	//                      missing-parent check looks at projects' parents)
	parentsMap := map[string]parser.NodeHelper{
		"roots.root.test.com": {
			RestName:    "root.Root",
			IsSingleton: true,
		},
		"orgs.org.test.com": {
			RestName: "orgs.Org",
			Parents:  []string{"roots.root.test.com"},
		},
		"projects.project.test.com": {
			RestName: "projects.Project",
			Parents:  []string{"roots.root.test.com", "orgs.org.test.com"},
		},
	}

	tests := []struct {
		name       string
		uri        string
		wantMissed []string
	}{
		{
			name:       "canonical tokens — parent present",
			uri:        "/v1/orgs/{orgs.Org}/projects/{projects.Project}",
			wantMissed: nil,
		},
		{
			name:       "alias tokens — parent present",
			uri:        "/v1/orgs/{org}/projects/{project}",
			wantMissed: nil,
		},
		{
			name:       "mixed alias/canonical — parent present",
			uri:        "/v1/orgs/{org}/projects/{projects.Project}",
			wantMissed: nil,
		},
		{
			name:       "missing non-singleton parent",
			uri:        "/v1/projects/{project}",
			wantMissed: []string{"orgs.Org"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			missing, _ := parser.ValidateRequiredParents(tc.uri, "projects.project.test.com", parentsMap, nil)
			if len(missing) != len(tc.wantMissed) {
				t.Fatalf("missing = %v, want %v", missing, tc.wantMissed)
			}
			for i := range missing {
				if missing[i] != tc.wantMissed[i] {
					t.Errorf("missing[%d] = %q, want %q", i, missing[i], tc.wantMissed[i])
				}
			}
		})
	}
}
