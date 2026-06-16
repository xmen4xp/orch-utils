// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"net/http"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// makeSchema returns a minimal JSONSchemaProps suitable as
// NodeInfo.Schema for the build tests below.
func makeSchema() *apiextensionsv1.JSONSchemaProps {
	return &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"spec": {
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"name": {Type: "string"},
				},
			},
			"status": {
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"ready": {Type: "boolean"},
				},
			},
		},
	}
}

// methodSet is a tiny helper for building a Methods set.
func methodSet(methods ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		out[m] = struct{}{}
	}
	return out
}

func TestBuild_EmptyDomain(t *testing.T) {
	b := New()
	spec := b.Build("hd.cisco.com")
	if spec == nil {
		t.Fatal("Build returned nil for empty domain")
	}
	if spec.Paths == nil {
		t.Fatal("spec.Paths is nil")
	}
	if len(spec.Tags) != 0 {
		t.Errorf("expected no tags, got %d", len(spec.Tags))
	}
}

func TestBuild_BasicCRUD(t *testing.T) {
	b := New()
	b.AddCRDNode("hd.cisco.com", "orgs.orgs.hd.cisco.com",
		NodeInfo{
			Name:        "orgs.Org",
			Description: "Organization",
			Schema:      makeSchema(),
		},
		[]RestURIs{{
			URI:       "/v1/organizations/{orgs.Org}",
			Methods:   methodSet(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete),
			TypeOfURI: DefaultURI,
		}},
	)

	spec := b.Build("hd.cisco.com")

	pi := spec.Paths.Value("/v1/organizations/{orgs.Org}")
	if pi == nil {
		t.Fatal("expected /v1/organizations/{orgs.Org} path")
	}
	if pi.Get == nil || pi.Get.OperationID != "getOrg" {
		t.Errorf("expected getOrg, got %+v", pi.Get)
	}
	if pi.Put == nil || pi.Put.OperationID != "putOrg" {
		t.Errorf("expected putOrg, got %+v", pi.Put)
	}
	if pi.Patch == nil || pi.Patch.OperationID != "patchOrg" {
		t.Errorf("expected patchOrg, got %+v", pi.Patch)
	}
	if pi.Delete == nil || pi.Delete.OperationID != "deleteOrg" {
		t.Errorf("expected deleteOrg, got %+v", pi.Delete)
	}

	// Tag emitted from CRD description.
	if len(spec.Tags) != 1 || spec.Tags[0].Name != "Org" {
		t.Errorf("expected single Org tag, got %+v", spec.Tags)
	}
}

// TestBuild_PerDomainCollisionScoping is the regression test for
// Issue 2. The same Kind "User" is registered in two different
// domains. Each domain's spec emits a BARE `User` tag — no
// cross-contamination, no over-qualification.
func TestBuild_PerDomainCollisionScoping(t *testing.T) {
	b := New()
	b.AddCRDNode("hd.cisco.com", "users.users.hd.cisco.com",
		NodeInfo{Name: "users.User", Description: "Application user", Schema: makeSchema()},
		[]RestURIs{{URI: "/v1/users/{users.User}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
	)
	b.AddCRDNode("admin.nexus.com", "users.user.admin.nexus.com",
		NodeInfo{Name: "user.User", Description: "Admin user", Schema: makeSchema()},
		[]RestURIs{{URI: "/v1/users/{user.User}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
	)

	hd := b.Build("hd.cisco.com")
	if hd.Paths.Value("/v1/users/{users.User}").Get.OperationID != "getUser" {
		t.Errorf("hd.cisco.com: expected getUser, got %q", hd.Paths.Value("/v1/users/{users.User}").Get.OperationID)
	}
	if len(hd.Tags) != 1 || hd.Tags[0].Name != "User" {
		t.Errorf("hd.cisco.com: expected single bare User tag, got %+v", hd.Tags)
	}

	admin := b.Build("admin.nexus.com")
	if admin.Paths.Value("/v1/users/{user.User}").Get.OperationID != "getUser" {
		t.Errorf("admin.nexus.com: expected getUser, got %q", admin.Paths.Value("/v1/users/{user.User}").Get.OperationID)
	}
	if len(admin.Tags) != 1 || admin.Tags[0].Name != "User" {
		t.Errorf("admin.nexus.com: expected single bare User tag, got %+v", admin.Tags)
	}
}

// TestBuild_SameDomainCollisionQualifies asserts the inverse of the
// previous test: when two Kinds collide WITHIN a domain, both get the
// package-prefix qualification. Regression coverage for ensuring our
// per-domain scoping didn't accidentally turn collision detection off.
func TestBuild_SameDomainCollisionQualifies(t *testing.T) {
	b := New()
	b.AddCRDNode("hd.cisco.com", "aislices.aislice.hd.cisco.com",
		NodeInfo{Name: "aislice.AISlice", Description: "Config AISlice", Schema: makeSchema()},
		[]RestURIs{{URI: "/v1/config/aislices/{aislice.AISlice}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
	)
	b.AddCRDNode("hd.cisco.com", "aislices.discoveredaislice.hd.cisco.com",
		NodeInfo{Name: "discoveredaislice.AISlice", Description: "Inventory AISlice", Schema: makeSchema()},
		[]RestURIs{{URI: "/v1/inventory/aislices/{discoveredaislice.AISlice}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
	)

	spec := b.Build("hd.cisco.com")
	if op := spec.Paths.Value("/v1/config/aislices/{aislice.AISlice}").Get; op == nil || op.OperationID != "getAisliceAISlice" {
		t.Errorf("expected getAisliceAISlice, got %+v", op)
	}
	if op := spec.Paths.Value("/v1/inventory/aislices/{discoveredaislice.AISlice}").Get; op == nil || op.OperationID != "getDiscoveredaisliceAISlice" {
		t.Errorf("expected getDiscoveredaisliceAISlice, got %+v", op)
	}

	// Two qualified tags, no bare "AISlice".
	gotTags := map[string]bool{}
	for _, tag := range spec.Tags {
		gotTags[tag.Name] = true
	}
	if gotTags["AISlice"] {
		t.Error("unexpected bare AISlice tag")
	}
	for _, want := range []string{"AisliceAISlice", "DiscoveredaisliceAISlice"} {
		if !gotTags[want] {
			t.Errorf("missing expected tag %q", want)
		}
	}
}

// TestBuild_SnapshotDeterminism is the regression test for Issue 1.
// The same input set added in different orders, or split across
// multiple Reset/Add cycles, must produce byte-identical JSON output.
// This is the structural guarantee that the orphan-tag bug (and any
// similar order-dependent emission bug) cannot recur.
func TestBuild_SnapshotDeterminism(t *testing.T) {
	// Define a canonical input set: two AISlice CRDs (force a
	// collision) + one Org.
	type input struct {
		crdType string
		info    NodeInfo
		uris    []RestURIs
	}
	inputs := []input{
		{
			"orgs.orgs.hd.cisco.com",
			NodeInfo{Name: "orgs.Org", Description: "Org", Schema: makeSchema()},
			[]RestURIs{{URI: "/v1/orgs/{orgs.Org}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
		},
		{
			"aislices.aislice.hd.cisco.com",
			NodeInfo{Name: "aislice.AISlice", Description: "Config AISlice", Schema: makeSchema()},
			[]RestURIs{{URI: "/v1/aislices/{aislice.AISlice}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
		},
		{
			"aislices.discoveredaislice.hd.cisco.com",
			NodeInfo{Name: "discoveredaislice.AISlice", Description: "Inventory AISlice", Schema: makeSchema()},
			[]RestURIs{{URI: "/v1/discovered/{discoveredaislice.AISlice}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
		},
	}

	specForOrder := func(order []int) []byte {
		b := New()
		for _, i := range order {
			b.AddCRDNode("hd.cisco.com", inputs[i].crdType, inputs[i].info, inputs[i].uris)
		}
		raw, err := json.Marshal(b.Build("hd.cisco.com"))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return raw
	}

	baseline := specForOrder([]int{0, 1, 2})

	// Try every permutation.
	for _, order := range [][]int{
		{0, 1, 2},
		{0, 2, 1},
		{1, 0, 2},
		{1, 2, 0},
		{2, 0, 1},
		{2, 1, 0},
	} {
		got := specForOrder(order)
		if string(got) != string(baseline) {
			t.Errorf("Build is not deterministic across insertion orders\nbaseline order [0 1 2]\nfailing  order %v", order)
		}
	}

	// Reset + re-add must produce the same spec.
	b := New()
	for _, i := range []int{0, 1, 2} {
		b.AddCRDNode("hd.cisco.com", inputs[i].crdType, inputs[i].info, inputs[i].uris)
	}
	first := mustMarshal(t, b.Build("hd.cisco.com"))

	// Add a noise node, then remove it. Final state matches the
	// initial set — Build output must therefore be identical to
	// `first`. This is the regression test for the api-gw orphan
	// tag bug: any path that left lingering state behind would
	// break this.
	b.AddCRDNode("hd.cisco.com", "noise.crd",
		NodeInfo{Name: "noise.Noise", Description: "noise", Schema: makeSchema()},
		[]RestURIs{{URI: "/v1/noise", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI}},
	)
	b.RemoveCRDNode("hd.cisco.com", "noise.crd")
	second := mustMarshal(t, b.Build("hd.cisco.com"))

	if string(first) != string(second) {
		t.Errorf("Add + Remove leaked state into Build output")
	}
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

func TestBuild_StatusURIEmitsAllVerbs(t *testing.T) {
	b := New()
	b.AddCRDNode("hd.cisco.com", "x.crd",
		NodeInfo{Name: "x.X", Description: "X", Schema: makeSchema()},
		[]RestURIs{
			{URI: "/v1/x/{x.X}", Methods: methodSet(http.MethodGet), TypeOfURI: DefaultURI},
			{URI: "/v1/x/{x.X}/status", Methods: methodSet(http.MethodGet, http.MethodPut, http.MethodPatch), TypeOfURI: StatusURI},
		},
	)
	spec := b.Build("hd.cisco.com")
	pi := spec.Paths.Value("/v1/x/{x.X}/status")
	if pi == nil {
		t.Fatal("missing /status path")
	}
	if pi.Get == nil || pi.Get.OperationID != "getXStatus" {
		t.Errorf("expected getXStatus")
	}
	if pi.Put == nil || pi.Put.OperationID != "putXStatus" {
		t.Errorf("expected putXStatus")
	}
	if pi.Patch == nil || pi.Patch.OperationID != "patchXStatus" {
		t.Errorf("expected patchXStatus")
	}
}

func TestBuild_Extension(t *testing.T) {
	b := New()
	b.AddExtension("hd.cisco.com", ExtensionSpec{
		URI: "/v1/metrics",
		OpenAPIPathSpec: `get:
  operationId: getMetrics
  tags: ["Metrics"]
  responses:
    "200":
      description: ok
`,
	})
	spec := b.Build("hd.cisco.com")
	pi := spec.Paths.Value("/v1/metrics")
	if pi == nil || pi.Get == nil {
		t.Fatal("extension /v1/metrics not installed")
	}
	if pi.Get.OperationID != "getMetrics" {
		t.Errorf("expected getMetrics, got %q", pi.Get.OperationID)
	}
}

// TestBuild_AddCRDNodeMergesByURI verifies that calling AddCRDNode
// twice for the same (domain, crdType) merges the URI lists by URI
// string rather than replacing. This is the contract that lets
// incremental adapters (openapi-generator) walk one URI at a time
// without maintaining a caller-side aggregator. Each URI's original
// methods (GET / LIST / PUT) are preserved verbatim across calls.
func TestBuild_AddCRDNodeMergesByURI(t *testing.T) {
	b := New()
	info := NodeInfo{Name: "orgs.Org", Description: "Org", Schema: makeSchema()}

	// First call — collection URI with LIST.
	b.AddCRDNode("hd.cisco.com", "orgs.crd", info, []RestURIs{
		{URI: "/v1/organizations", Methods: map[string]struct{}{"LIST": {}}},
	})
	// Second call — single-item URI with GET / PUT / DELETE. Must
	// NOT erase the prior /v1/organizations entry.
	b.AddCRDNode("hd.cisco.com", "orgs.crd", info, []RestURIs{
		{URI: "/v1/organizations/{org}", Methods: map[string]struct{}{
			"GET": {}, "PUT": {}, "DELETE": {},
		}},
	})

	spec := b.Build("hd.cisco.com")
	listOp := spec.Paths.Value("/v1/organizations")
	if listOp == nil || listOp.Get == nil {
		t.Fatal("/v1/organizations operation missing after second AddCRDNode")
	}
	if got := listOp.Get.OperationID; got != "listOrgs" {
		t.Errorf("expected listOrgs (LIST preserved), got %q", got)
	}
	itemOp := spec.Paths.Value("/v1/organizations/{org}")
	if itemOp == nil || itemOp.Get == nil || itemOp.Put == nil {
		t.Fatal("/v1/organizations/{org} operations missing")
	}
	if got := itemOp.Get.OperationID; got != "getOrg" {
		t.Errorf("expected getOrg, got %q", got)
	}
}

// TestBuild_OrphanTagPruning verifies that a CRD declared with a
// Description but no registered URIs does not produce a root-level
// tag entry (which would have no operation referencing it). Regression
// coverage for the orphan-tag defect surfaced by the e2e test
// `declares no orphan tags`.
func TestBuild_OrphanTagPruning(t *testing.T) {
	b := New()
	// Node A: registered with a URI — its tag MUST be emitted.
	b.AddCRDNode("hd.cisco.com", "a.crd",
		NodeInfo{Name: "pkg.Used", Description: "Used CRD", Schema: makeSchema()},
		[]RestURIs{{URI: "/v1/used", Methods: map[string]struct{}{"GET": {}}}})
	// Node B: registered with NO URIs — its tag would be orphan
	// without pruning.
	b.AddCRDNode("hd.cisco.com", "b.crd",
		NodeInfo{Name: "pkg.Orphan", Description: "Orphan CRD", Schema: makeSchema()},
		nil)

	spec := b.Build("hd.cisco.com")

	got := map[string]bool{}
	for _, tag := range spec.Tags {
		got[tag.Name] = true
	}
	if !got["Used"] {
		t.Errorf("expected `Used` tag to be emitted (referenced by /v1/used), got %v", got)
	}
	if got["Orphan"] {
		t.Errorf("orphan tag `Orphan` was emitted but no operation references it: got %v", got)
	}
}

func TestBuild_ResetClearsAllDomains(t *testing.T) {
	b := New()
	b.AddCRDNode("hd.cisco.com", "x.crd", NodeInfo{Name: "x.X", Schema: makeSchema()}, nil)
	b.AddCRDNode("admin.nexus.com", "y.crd", NodeInfo{Name: "y.Y", Schema: makeSchema()}, nil)
	b.Reset()
	if doms := b.Domains(); len(doms) != 0 {
		t.Errorf("expected no domains after Reset, got %v", doms)
	}
}

func TestBuild_ResetDomain(t *testing.T) {
	b := New()
	b.AddCRDNode("hd.cisco.com", "x.crd", NodeInfo{Name: "x.X", Schema: makeSchema()}, nil)
	b.AddCRDNode("admin.nexus.com", "y.crd", NodeInfo{Name: "y.Y", Schema: makeSchema()}, nil)
	b.ResetDomain("hd.cisco.com")
	doms := b.Domains()
	if len(doms) != 1 || doms[0] != "admin.nexus.com" {
		t.Errorf("expected only admin.nexus.com after ResetDomain, got %v", doms)
	}
}
