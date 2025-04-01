// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package openapi

// Helper functions for validation rules.
func FloatPtr(i int) *float64 {
	val := float64(i)
	return &val
}

func IntPtr(i int) *int64 {
	val := int64(i)
	return &val
}
