// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"nexus-api-gw/pkg/model"
)

func TestQualifiedKind_NoCollision(t *testing.T) {
	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		"orgs.example.com":     {Name: "orgs.Org"},
		"projects.example.com": {Name: "projects.Project"},
	}
	defer func() { model.CrdTypeToNodeInfo = map[string]model.NodeInfo{} }()

	if got := qualifiedKind([]string{"orgs", "Org"}); got != "Org" {
		t.Errorf("expected Org (no collision), got %q", got)
	}
}

func TestQualifiedKind_WithCollision(t *testing.T) {
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
	if got := qualifiedKind([]string{"orgs", "Org"}); got != "Org" {
		t.Errorf("expected Org (no collision), got %q", got)
	}
}

func TestQualifiedKind_NoKindPortion(t *testing.T) {
	if got := qualifiedKind([]string{"justpackage"}); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
