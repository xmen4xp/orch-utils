// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package structtype

// +k8s:openapi-gen=true
type AtomicStruct struct {
	// +structType=atomic
	Field      ContainedStruct
	OtherField int
}

// +k8s:openapi-gen=true
type ContainedStruct struct{}

// +k8s:openapi-gen=true
// +structType=atomic
type DeclaredAtomicStruct struct {
	Field int
}
