// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import "strings"

// KindCounts is a per-domain histogram of Kind occurrences (the portion
// of NodeInfo.Name after the first dot). It is the input to collision
// detection — when count > 1, the Kind needs a package prefix to be
// unambiguous within its domain.
//
// Callers compute KindCounts once per Build invocation, from the
// snapshot of NodeInfos for the target domain only. This is what makes
// collision detection per-domain: a Kind shared across two different
// domains will appear once in each domain's KindCounts and therefore
// not be marked as colliding.
type KindCounts map[string]int

// isCollidingKind reports whether kind appears more than once in the
// given per-domain KindCounts. Returns false for the empty kind.
func isCollidingKind(kind string, counts KindCounts) bool {
	if kind == "" {
		return false
	}
	return counts[kind] > 1
}

// qualifiedKind returns the Kind portion of a CRD's NodeInfo.Name (the
// part after the first dot), prefixed with a Title-cased package
// segment iff more than one CRD in the same domain shares that Kind.
// Returns "" when nameParts has no Kind portion.
//
// Examples (within a single domain):
//
//	["orgs", "Org"]                       (no collision) -> "Org"
//	["aislice", "AISlice"]                (collision)    -> "AisliceAISlice"
//	["discoveredaislice", "AISlice"]      (collision)    -> "DiscoveredaisliceAISlice"
func qualifiedKind(nameParts []string, counts KindCounts) string {
	if len(nameParts) < 2 {
		return ""
	}
	kind := nameParts[1]
	if isCollidingKind(kind, counts) {
		return titleFirst(nameParts[0]) + kind
	}
	return kind
}

// titleFirst upper-cases the first rune of s. Used instead of the
// deprecated strings.Title for our simple ASCII package-name inputs.
func titleFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// kindPlural returns the plural form of a Kind name, derived solely
// from the Kind itself. If the Kind already ends in "s" (e.g.
// "Clusters", "DataCenters", "Nodes"), it is returned unchanged.
// Otherwise an "s" is appended (e.g. "Org" -> "Orgs",
// "AISlice" -> "AISlices").
func kindPlural(kind string) string {
	if strings.HasSuffix(kind, "s") {
		return kind
	}
	return kind + "s"
}

// lastStaticSegment returns the trailing non-parameter segment of a
// URI, or "" if the URI ends in a path parameter "{...}". Used only
// for link traversal URIs where the trailing segment is the Go field
// name being navigated to.
func lastStaticSegment(uri string) string {
	parts := strings.Split(strings.TrimRight(uri, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if strings.HasPrefix(last, "{") {
		return ""
	}
	return last
}

// kindCountsFromNodes computes a per-domain Kind histogram from a
// snapshot of NodeInfos. Used internally by Build; exported under the
// `BuildKindCounts` name for test convenience.
func kindCountsFromNodes(nodes []NodeInfo) KindCounts {
	counts := make(KindCounts, len(nodes))
	for _, n := range nodes {
		parts := strings.SplitN(n.Name, ".", 2)
		if len(parts) != 2 {
			continue
		}
		counts[parts[1]]++
	}
	return counts
}

// BuildKindCounts is the exported wrapper around kindCountsFromNodes
// for use in adapter integration tests. Production code computes
// counts internally; this entry point is provided so adapter tests
// can validate the per-domain scoping without reaching into private
// builder state.
func BuildKindCounts(nodes []NodeInfo) KindCounts {
	return kindCountsFromNodes(nodes)
}
