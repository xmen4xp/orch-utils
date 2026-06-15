// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"reflect"
	"regexp"
	"testing"
)

func TestResolveUriParams(t *testing.T) {
	r := regexp.MustCompile(`{([^{}]+)}`)

	tests := []struct {
		name       string
		uri        string
		pathParams map[string]string
		wantTokens []string
	}{
		{
			name:       "nil PathParams passes canonical through",
			uri:        "/v1/organizations/{orgs.Org}/projects/{projects.Project}",
			pathParams: nil,
			wantTokens: []string{"orgs.Org", "projects.Project"},
		},
		{
			name: "alias tokens resolved to canonical",
			uri:  "/v1/organizations/{org}/projects/{project}",
			pathParams: map[string]string{
				"org":     "orgs.Org",
				"project": "projects.Project",
			},
			wantTokens: []string{"orgs.Org", "projects.Project"},
		},
		{
			name:       "unmapped token preserved",
			uri:        "/v1/aislices/{aislice.AISlice}",
			pathParams: map[string]string{"org": "orgs.Org"},
			wantTokens: []string{"aislice.AISlice"},
		},
		{
			name:       "no tokens in URI",
			uri:        "/v1/organizations",
			pathParams: nil,
			wantTokens: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := r.FindAllStringSubmatch(tt.uri, -1)
			resolved := resolveUriParams(raw, tt.pathParams)
			var gotTokens []string
			for _, p := range resolved {
				if len(p) >= 2 {
					gotTokens = append(gotTokens, p[1])
				}
			}
			if !reflect.DeepEqual(gotTokens, tt.wantTokens) {
				t.Errorf("resolveUriParams tokens = %v, want %v", gotTokens, tt.wantTokens)
			}
		})
	}
}
