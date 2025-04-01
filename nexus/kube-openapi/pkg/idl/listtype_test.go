// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package idl_test

// This example shows how to use the listType map attribute and how to
// specify a key to identify elements of the list. The listMapKey
// attribute is used to specify that Name is the key of the map.
func ExampleListType_map() {
	type SomeStruct struct {
		Name  string
		Value string
	}
	type SomeAPI struct {
		// +listType=map
		// +listMapKey=name
		elements []SomeStruct
	}
}

// This example shows how to use the listType set attribute to specify
// that this list should be treated as a set: items in the list can't be
// duplicated.
func ExampleListType_set() {
	type SomeAPI struct {
		// +listType=set
		keys []string
	}
}

// This example shows how to use the listType atomic attribute to
// specify that this list should be treated as a whole.
func ExampleListType_atomic() {
	type SomeStruct struct {
		Name  string
		Value string
	}

	type SomeAPI struct {
		// +listType=atomic
		elements []SomeStruct
	}
}
