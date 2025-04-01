// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package idl_test

// This example shows how to use the structType atomic attribute to
// specify that this struct should be treated as a whole.
func ExampleStructType_atomic() {
	type SomeStruct struct {
		Name  string
		Value string
	}
	type SomeAPI struct {
		// +structType=atomic
		elements SomeStruct
	}
}
