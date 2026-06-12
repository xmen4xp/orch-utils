// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"nexus/openapi-generator/pkg/model"
)

func TestQualifiedKind_NoCollision(t *testing.T) {
	// Reset and seed a single CRD per Kind.
	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		"orgs.example.com":     {Name: "orgs.Org"},
		"projects.example.com": {Name: "projects.Project"},
	}
	defer func() { model.CrdTypeToNodeInfo = map[string]model.NodeInfo{} }()

	if got := qualifiedKind([]string{"orgs", "Org"}); got != "Org" {
		t.Errorf("expected Org (no collision), got %q", got)
	}
	if got := qualifiedKind([]string{"projects", "Project"}); got != "Project" {
		t.Errorf("expected Project (no collision), got %q", got)
	}
}

func TestQualifiedKind_WithCollision(t *testing.T) {
	// Two CRDs sharing Kind "AISlice" under different packages.
	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		"aislices.example.com":           {Name: "aislice.AISlice"},
		"discoveredaislices.example.com": {Name: "discoveredaislice.AISlice"},
		"orgs.example.com":               {Name: "orgs.Org"},
	}
	defer func() { model.CrdTypeToNodeInfo = map[string]model.NodeInfo{} }()

	if got := qualifiedKind([]string{"aislice", "AISlice"}); got != "AisliceAISlice" {
		t.Errorf("expected AisliceAISlice (collision -> package prefix), got %q", got)
	}
	if got := qualifiedKind([]string{"discoveredaislice", "AISlice"}); got != "DiscoveredaisliceAISlice" {
		t.Errorf("expected DiscoveredaisliceAISlice (collision -> package prefix), got %q", got)
	}
	// Non-colliding kind in the same map stays untouched.
	if got := qualifiedKind([]string{"orgs", "Org"}); got != "Org" {
		t.Errorf("expected Org (no collision), got %q", got)
	}
}

func TestQualifiedKind_NoKindPortion(t *testing.T) {
	if got := qualifiedKind([]string{"justpackage"}); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
