// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package idl_test

// This example shows how to use the mapType atomic attribute to
// specify that this map should be treated as a whole.
func ExampleMapType_atomic() {
	type SomeAPI struct {
		// +mapType=atomic
		elements map[string]string
	}
}
