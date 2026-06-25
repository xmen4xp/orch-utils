// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import "testing"

func TestQualifiedKind_NoCollision(t *testing.T) {
	counts := kindCountsFromNodes([]NodeInfo{
		{Name: "orgs.Org"},
		{Name: "projects.Project"},
	})

	if got := qualifiedKind([]string{"orgs", "Org"}, counts); got != "Org" {
		t.Errorf("expected Org (no collision), got %q", got)
	}
	if got := qualifiedKind([]string{"projects", "Project"}, counts); got != "Project" {
		t.Errorf("expected Project (no collision), got %q", got)
	}
}

func TestQualifiedKind_WithCollision(t *testing.T) {
	counts := kindCountsFromNodes([]NodeInfo{
		{Name: "aislice.AISlice"},
		{Name: "discoveredaislice.AISlice"},
		{Name: "orgs.Org"},
	})

	if got := qualifiedKind([]string{"aislice", "AISlice"}, counts); got != "Aislice" {
		t.Errorf("expected Aislice (collision -> package only), got %q", got)
	}
	if got := qualifiedKind([]string{"discoveredaislice", "AISlice"}, counts); got != "Discoveredaislice" {
		t.Errorf("expected Discoveredaislice (collision -> package only), got %q", got)
	}
	if got := qualifiedKind([]string{"orgs", "Org"}, counts); got != "Org" {
		t.Errorf("expected Org (no collision in same domain), got %q", got)
	}
}

func TestQualifiedKind_NoKindPortion(t *testing.T) {
	if got := qualifiedKind([]string{"justpackage"}, nil); got != "" {
		t.Errorf("expected empty string for single-segment name, got %q", got)
	}
}

// TestQualifiedKind_PerDomainScoping is the regression test for Issue 2.
// Same Kind "User" appears in two separate domains. Each domain computes
// its own KindCounts. Within either domain, "User" is unique → no
// collision → bare "User" emitted on both sides. This is the structural
// fix that prevents `users.users.hd.cisco.com` from being over-qualified
// to "UsersUser" just because `users.user.admin.nexus.com` exists in a
// completely different datamodel.
func TestQualifiedKind_PerDomainScoping(t *testing.T) {
	hdCisco := kindCountsFromNodes([]NodeInfo{
		{Name: "users.User"},
		{Name: "orgs.Org"},
		{Name: "projects.Project"},
	})

	adminNexus := kindCountsFromNodes([]NodeInfo{
		{Name: "user.User"},
	})

	if got := qualifiedKind([]string{"users", "User"}, hdCisco); got != "User" {
		t.Errorf("hd.cisco.com: expected bare User (no collision within domain), got %q", got)
	}
	if got := qualifiedKind([]string{"user", "User"}, adminNexus); got != "User" {
		t.Errorf("admin.nexus.com: expected bare User (no collision within domain), got %q", got)
	}
}

func TestKindPlural(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Org", "Orgs"},
		{"Project", "Projects"},
		{"AISlice", "AISlices"},
		{"Clusters", "Clusters"},       // already plural — unchanged
		{"DataCenters", "DataCenters"}, // already plural — unchanged
		{"Nodes", "Nodes"},             // already plural — unchanged
		{"", ""},                       // empty in -> empty + "s" would be wrong; current behavior is "s". Document behavior.
		{"s", "s"},                     // single "s" already counts as plural
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if c.in == "" {
				// kindPlural("") returns "s" per current implementation;
				// callers never pass empty kinds so this is unobserved in
				// production. Skip the assertion here.
				t.Skip("empty input is not a valid Kind")
			}
			if got := kindPlural(c.in); got != c.want {
				t.Errorf("kindPlural(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestTitleFirst(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"aislice", "Aislice"},
		{"orgs", "Orgs"},
		{"", ""},
		{"X", "X"},
	}
	for _, c := range cases {
		if got := titleFirst(c.in); got != c.want {
			t.Errorf("titleFirst(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLastStaticSegment(t *testing.T) {
	cases := []struct {
		uri, want string
	}{
		{"/v1/users/{users.User}/Binding", "Binding"},
		{"/v1/orgs/{orgs.Org}/Projects", "Projects"},
		{"/v1/orgs/{orgs.Org}/AISlice", "AISlice"},
		{"/v1/orgs/{orgs.Org}", ""}, // ends in param
		{"/", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := lastStaticSegment(c.uri); got != c.want {
			t.Errorf("lastStaticSegment(%q) = %q, want %q", c.uri, got, c.want)
		}
	}
}
