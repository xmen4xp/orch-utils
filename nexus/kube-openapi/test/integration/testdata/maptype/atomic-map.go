// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package maptype

// +k8s:openapi-gen=true
type AtomicMap struct {
	// +mapType=atomic
	KeyValue map[string]string
}
