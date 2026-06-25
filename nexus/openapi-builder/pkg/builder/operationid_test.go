// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import "testing"

// TestGetOperationID covers the full operationId derivation table:
// LIST / GET / PUT / PATCH / DELETE on default and status URIs, plus
// link traversal (SingleLink and NamedLink). No collisions exercised
// here — see the collision sub-test below.
func TestGetOperationID(t *testing.T) {
	// Single-domain snapshot with no Kind collisions.
	counts := kindCountsFromNodes([]NodeInfo{
		{Name: "orgs.Org"},
		{Name: "projects.Project"},
		{Name: "aislice.AISlice"},
		{Name: "users.User"},
		{Name: "spaces.Space"},
	})

	cases := []struct {
		name    string
		method  string
		uri     string
		info    NodeInfo
		uriType URIType
		want    string
	}{
		{"list AISlice", "LIST", "/v1/config/spaces/{spaces.Space}/aislices", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "listAISlices"},
		{"list Org", "LIST", "/v1/organizations", NodeInfo{Name: "orgs.Org"}, DefaultURI, "listOrgs"},
		{"get AISlice", "GET", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "getAISlice"},
		{"put AISlice", "PUT", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "putAISlice"},
		{"patch AISlice", "PATCH", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "patchAISlice"},
		{"delete AISlice", "DELETE", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "deleteAISlice"},
		{"status get", "GET", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status", NodeInfo{Name: "aislice.AISlice"}, StatusURI, "getAISliceStatus"},
		{"status put", "PUT", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status", NodeInfo{Name: "aislice.AISlice"}, StatusURI, "putAISliceStatus"},
		{"status patch", "PATCH", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status", NodeInfo{Name: "aislice.AISlice"}, StatusURI, "patchAISliceStatus"},
		{"single-link Space.AISlice", "GET", "/v1/config/spaces/{spaces.Space}/AISlice", NodeInfo{Name: "spaces.Space"}, SingleLinkURI, "getSpaceAISlice"},
		{"single-link User.Binding", "GET", "/v1/users/{users.User}/Binding", NodeInfo{Name: "users.User"}, SingleLinkURI, "getUserBinding"},
		{"named-link Org.Projects", "GET", "/v1/organizations/{orgs.Org}/Projects", NodeInfo{Name: "orgs.Org"}, NamedLinkURI, "getOrgProjects"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := GetOperationID(c.method, c.uri, c.info, c.uriType, counts); got != c.want {
				t.Errorf("GetOperationID(%q, %q, %q, %v) = %q, want %q",
					c.method, c.uri, c.info.Name, c.uriType, got, c.want)
			}
		})
	}
}

// TestGetOperationID_WithCollision asserts that operationIds for
// colliding Kinds receive the package prefix. Regression coverage
// for Issue 1's family of bugs: any drift in collision propagation
// must break this test.
func TestGetOperationID_WithCollision(t *testing.T) {
	counts := kindCountsFromNodes([]NodeInfo{
		{Name: "aislice.AISlice"},
		{Name: "discoveredaislice.AISlice"},
		{Name: "orgs.Org"},
	})

	cases := []struct {
		name    string
		method  string
		uri     string
		info    NodeInfo
		uriType URIType
		want    string
	}{
		{"get AISlice (config)", "GET", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "getAislice"},
		{"put AISlice (config)", "PUT", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "putAislice"},
		{"get AISlice (discovered)", "GET", "/v1/inventory/datacenters/{datacenter}/clusters/{cluster}/aislices/{discoveredaislice.AISlice}", NodeInfo{Name: "discoveredaislice.AISlice"}, DefaultURI, "getDiscoveredaislice"},
		{"status put AISlice (config)", "PUT", "/v1/config/.../{aislice.AISlice}/status", NodeInfo{Name: "aislice.AISlice"}, StatusURI, "putAisliceStatus"},
		{"list AISlices (config)", "LIST", "/v1/config/spaces/{spaces.Space}/aislices", NodeInfo{Name: "aislice.AISlice"}, DefaultURI, "listAislices"},
		{"get Org (no collision)", "GET", "/v1/organizations/{orgs.Org}", NodeInfo{Name: "orgs.Org"}, DefaultURI, "getOrg"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := GetOperationID(c.method, c.uri, c.info, c.uriType, counts); got != c.want {
				t.Errorf("GetOperationID(%q, %q, %q, %v) = %q, want %q",
					c.method, c.uri, c.info.Name, c.uriType, got, c.want)
			}
		})
	}
}
