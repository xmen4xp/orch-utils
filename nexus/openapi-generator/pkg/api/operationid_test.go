// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"nexus/openapi-generator/pkg/model"
)

func TestGetOperationID(t *testing.T) {
	// Seed the URI -> URIInfo map used by getOperationID for sub-URI classification.
	model.UriToUriInfo = map[string]model.RestURIInfo{
		"/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}":        {TypeOfURI: model.DefaultURI},
		"/v1/config/spaces/{spaces.Space}/aislices":                          {TypeOfURI: model.DefaultURI},
		"/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status": {TypeOfURI: model.StatusURI},
		"/v1/config/spaces/{spaces.Space}/AISlice":                           {TypeOfURI: model.SingleLinkURI},
		"/v1/organizations/{orgs.Org}":                                       {TypeOfURI: model.DefaultURI},
		"/v1/organizations":                                                  {TypeOfURI: model.DefaultURI},
		"/v1/organizations/{orgs.Org}/Projects":                              {TypeOfURI: model.NamedLinkURI},
		"/v1/users/{users.User}/Binding":                                     {TypeOfURI: model.SingleLinkURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}":                {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters":                                          {TypeOfURI: model.DefaultURI},
	}
	t.Cleanup(func() { model.UriToUriInfo = map[string]model.RestURIInfo{} })

	tests := []struct {
		name   string
		method string
		uri    string
		kind   string // crdInfo.Name's <group>.<Kind>
		want   string
	}{
		{"list AISlice", "LIST", "/v1/config/spaces/{spaces.Space}/aislices", "aislice.AISlice", "getAISlices"},
		{"get AISlice", "GET", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "getAISlice"},
		{"put AISlice", "PUT", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "putAISlice"},
		{"patch AISlice", "PATCH", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "patchAISlice"},
		{"delete AISlice", "DELETE", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "deleteAISlice"},
		{"status get", "GET", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status", "aislice.AISlice", "getAISliceStatus"},
		{"status put", "PUT", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status", "aislice.AISlice", "putAISliceStatus"},
		{"single-link Space.AISlice", "GET", "/v1/config/spaces/{spaces.Space}/AISlice", "spaces.Space", "getSpaceAISlice"},
		{"single-link User.Binding", "GET", "/v1/users/{users.User}/Binding", "users.User", "getUserBinding"},
		{"named-link Org.Projects", "GET", "/v1/organizations/{orgs.Org}/Projects", "orgs.Org", "getOrgProjects"},
		{"list Org -> organizations", "LIST", "/v1/organizations", "orgs.Org", "getOrganizations"},
		{"get Org", "GET", "/v1/organizations/{orgs.Org}", "orgs.Org", "getOrg"},
		{"list DataCenters (plural-named Kind)", "LIST", "/v1/inventory/datacenters", "datacenters.DataCenters", "getDataCenters"},
		{"get DataCenters", "GET", "/v1/inventory/datacenters/{datacenters.DataCenters}", "datacenters.DataCenters", "getDataCenters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := model.NodeInfo{Name: tt.kind}
			got := getOperationID(tt.method, tt.uri, ci)
			if got != tt.want {
				t.Errorf("getOperationID(%q, %q) = %q; want %q", tt.method, tt.uri, got, tt.want)
			}
		})
	}
}
