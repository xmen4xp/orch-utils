// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package maptype

// +k8s:openapi-gen=true
type GranularMap struct {
	// +mapType=granular
	KeyValue map[string]string
}
