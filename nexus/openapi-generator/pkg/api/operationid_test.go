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
		// AISlice
		"/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}":        {TypeOfURI: model.DefaultURI},
		"/v1/config/spaces/{spaces.Space}/aislices":                          {TypeOfURI: model.DefaultURI},
		"/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status": {TypeOfURI: model.StatusURI},
		// Link traversal
		"/v1/config/spaces/{spaces.Space}/AISlice": {TypeOfURI: model.SingleLinkURI},
		"/v1/organizations/{orgs.Org}/Projects":    {TypeOfURI: model.NamedLinkURI},
		"/v1/users/{users.User}/Binding":           {TypeOfURI: model.SingleLinkURI},
		// Org
		"/v1/organizations/{orgs.Org}":          {TypeOfURI: model.DefaultURI},
		"/v1/organizations":                     {TypeOfURI: model.DefaultURI},
		"/v1/organizations/{orgs.Org}/projects": {TypeOfURI: model.DefaultURI},
		"/v1/organizations/{orgs.Org}/projects/{projects.Project}": {TypeOfURI: model.DefaultURI},
		// Plural-named Kinds (used to cause LIST/GET collisions under the old scheme)
		"/v1/inventory/datacenters/{datacenters.DataCenters}":                              {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters":                                                        {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}/clusters":                     {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}": {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}/nodes":               {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}/nodes/{nodes.Nodes}": {TypeOfURI: model.DefaultURI},
		// Same-segment, different Kind (ConfigCluster vs Clusters) — previously colliding
		"/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters":                                   {TypeOfURI: model.DefaultURI},
		"/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters/{configcluster.ConfigCluster}":     {TypeOfURI: model.DefaultURI},
		"/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters/{configcluster.ConfigCluster}/entitlements/{clusterentitlement.ClusterEntitlement}": {TypeOfURI: model.DefaultURI},
	}
	t.Cleanup(func() { model.UriToUriInfo = map[string]model.RestURIInfo{} })

	tests := []struct {
		name   string
		method string
		uri    string
		kind   string // crdInfo.Name's <group>.<Kind>
		want   string
	}{
		// Standard CRUD
		{"list AISlice", "LIST", "/v1/config/spaces/{spaces.Space}/aislices", "aislice.AISlice", "listAISlices"},
		{"get AISlice", "GET", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "getAISlice"},
		{"put AISlice", "PUT", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "putAISlice"},
		{"patch AISlice", "PATCH", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "patchAISlice"},
		{"delete AISlice", "DELETE", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}", "aislice.AISlice", "deleteAISlice"},
		// Status sub-URI
		{"status get", "GET", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status", "aislice.AISlice", "getAISliceStatus"},
		{"status put", "PUT", "/v1/config/spaces/{spaces.Space}/aislices/{aislice.AISlice}/status", "aislice.AISlice", "putAISliceStatus"},
		// Link traversal keeps parent + field name
		{"single-link Space.AISlice", "GET", "/v1/config/spaces/{spaces.Space}/AISlice", "spaces.Space", "getSpaceAISlice"},
		{"single-link User.Binding", "GET", "/v1/users/{users.User}/Binding", "users.User", "getUserBinding"},
		{"named-link Org.Projects", "GET", "/v1/organizations/{orgs.Org}/Projects", "orgs.Org", "getOrgProjects"},
		// Org (Kind-derived plural — URI segment is "organizations" but Kind is "Org" -> "Orgs")
		{"list Org", "LIST", "/v1/organizations", "orgs.Org", "listOrgs"},
		{"get Org", "GET", "/v1/organizations/{orgs.Org}", "orgs.Org", "getOrg"},
		{"list Project under Org", "LIST", "/v1/organizations/{orgs.Org}/projects", "projects.Project", "listProjects"},
		{"get Project under Org", "GET", "/v1/organizations/{orgs.Org}/projects/{projects.Project}", "projects.Project", "getProject"},

		// Plural-named Kinds — verb differentiation resolves LIST vs GET collisions
		{"list DataCenters (plural-named Kind)", "LIST", "/v1/inventory/datacenters", "datacenters.DataCenters", "listDataCenters"},
		{"get DataCenters", "GET", "/v1/inventory/datacenters/{datacenters.DataCenters}", "datacenters.DataCenters", "getDataCenters"},
		{"list Clusters (plural-named Kind)", "LIST", "/v1/inventory/datacenters/{datacenters.DataCenters}/clusters", "clusters.Clusters", "listClusters"},
		{"get Clusters", "GET", "/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}", "clusters.Clusters", "getClusters"},
		{"list Nodes (plural-named Kind)", "LIST", "/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}/nodes", "nodes.Nodes", "listNodes"},
		{"get Nodes", "GET", "/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}/nodes/{nodes.Nodes}", "nodes.Nodes", "getNodes"},

		// ConfigCluster vs Clusters disambiguation (previously both produced getClusters)
		{"list ConfigCluster (singular Kind)", "LIST", "/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters", "configcluster.ConfigCluster", "listConfigClusters"},
		{"get ConfigCluster", "GET", "/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters/{configcluster.ConfigCluster}", "configcluster.ConfigCluster", "getConfigCluster"},
		{"delete ClusterEntitlement", "DELETE", "/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters/{configcluster.ConfigCluster}/entitlements/{clusterentitlement.ClusterEntitlement}", "clusterentitlement.ClusterEntitlement", "deleteClusterEntitlement"},
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

// TestGetOperationID_NoCollisions exercises the same set of representative
// CRD URIs and asserts that every produced operationId is unique across the
// full table — guarding against future regressions in the naming scheme.
func TestGetOperationID_NoCollisions(t *testing.T) {
	model.UriToUriInfo = map[string]model.RestURIInfo{
		"/v1/inventory/datacenters":                           {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}": {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}/clusters":                     {TypeOfURI: model.DefaultURI},
		"/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}": {TypeOfURI: model.DefaultURI},
		"/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters":                                   {TypeOfURI: model.DefaultURI},
		"/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters/{configcluster.ConfigCluster}":     {TypeOfURI: model.DefaultURI},
	}
	t.Cleanup(func() { model.UriToUriInfo = map[string]model.RestURIInfo{} })

	cases := []struct {
		method, uri, kind string
	}{
		{"LIST", "/v1/inventory/datacenters", "datacenters.DataCenters"},
		{"GET", "/v1/inventory/datacenters/{datacenters.DataCenters}", "datacenters.DataCenters"},
		{"LIST", "/v1/inventory/datacenters/{datacenters.DataCenters}/clusters", "clusters.Clusters"},
		{"GET", "/v1/inventory/datacenters/{datacenters.DataCenters}/clusters/{clusters.Clusters}", "clusters.Clusters"},
		{"LIST", "/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters", "configcluster.ConfigCluster"},
		{"GET", "/v1/config/datacenters/{configdatacenter.ConfigDataCenter}/clusters/{configcluster.ConfigCluster}", "configcluster.ConfigCluster"},
	}

	seen := map[string]string{} // opId -> "method uri"
	for _, c := range cases {
		op := getOperationID(c.method, c.uri, model.NodeInfo{Name: c.kind})
		if prev, ok := seen[op]; ok {
			t.Errorf("operationId %q collides: %s %s vs %s", op, c.method, c.uri, prev)
		}
		seen[op] = c.method + " " + c.uri
	}
}
