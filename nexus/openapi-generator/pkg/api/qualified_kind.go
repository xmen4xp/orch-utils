// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"strings"

	"nexus/openapi-generator/pkg/model"
)

// qualifiedKind returns the Kind portion of a CRD's NodeInfo.Name
// (the part after the first dot), prefixed with a Title-cased package
// segment iff more than one CRD shares the same Kind. The compiler
// guarantees package.Kind uniqueness so the package prefix is always
// sufficient to disambiguate (e.g. aislice.AISlice vs
// discoveredaislice.AISlice -> "AisliceAISlice" / "DiscoveredaisliceAISlice").
// Returns "" when nameParts has no Kind portion.
func qualifiedKind(nameParts []string) string {
	if len(nameParts) < 2 {
		return ""
	}
	kind := nameParts[1]
	if model.IsCollidingKind(kind) {
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
