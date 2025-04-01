// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package pointer

// Of is a helper routine that allocates a new any value
// to store v and returns a pointer to it.
// Ref: https://github.com/xorcare/pointer?tab=readme-ov-file#a-little-copying-is-better-than-a-little-dependency
func Of[Value any](v Value) *Value {
	return &v
}
