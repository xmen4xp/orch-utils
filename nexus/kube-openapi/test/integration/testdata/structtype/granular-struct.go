// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package structtype

// +k8s:openapi-gen=true
type GranularStruct struct {
	// +structType=granular
	Field      ContainedStruct
	OtherField int
}
