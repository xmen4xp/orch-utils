// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package model

import "strings"

// IsCollidingKind reports whether more than one registered CRD shares the
// given Kind (the part after the dot in NodeInfo.Name). Used by OpenAPI
// emission to decide whether an operationId / tag needs a package prefix
// to avoid colliding with another CRD that has the same Kind under a
// different package (e.g. aislice.AISlice vs discoveredaislice.AISlice).
//
// The compiler guarantees package.Kind uniqueness, so the package prefix
// is always sufficient to disambiguate.
func IsCollidingKind(kind string) bool {
	if kind == "" {
		return false
	}
	crdTypeToNodeInfoMutex.Lock()
	defer crdTypeToNodeInfoMutex.Unlock()

	count := 0
	for _, ni := range CrdTypeToNodeInfo {
		parts := strings.SplitN(ni.Name, ".", 2)
		if len(parts) == 2 && parts[1] == kind {
			count++
			if count > 1 {
				return true
			}
		}
	}
	return false
}
