// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// The package is intended for testing the openapi-gen API rule
// checker. The API rule violations are in format of:
//
// `{rule-name},{package},{type},{(optional) field}`
//
// The checker should sort the violations before
// reporting to a file or stderr.
//
// We have the dummytype package separately from the listtype
// package to test the sorting behavior on package level, e.g.
//
//   -i "./testdata/listtype,./testdata/dummytype"
//   -i "./testdata/dummytype,./testdata/listtype"
//
// The violations from dummytype should always come first in
// report.

package dummytype

// +k8s:openapi-gen=true
type Foo struct {
	Second string
	First  int
}

// +k8s:openapi-gen=true
type Bar struct {
	ViolationBehind bool
	Violation       bool
}

// +k8s:openapi-gen=true
type Baz struct {
	Violation       bool
	ViolationBehind bool
}

// +k8s:openapi-gen=true
type StatusError struct {
	Code    int
	Message string
}
