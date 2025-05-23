// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package listtype

// +k8s:openapi-gen=true
type MapList struct {
	// +listType=map
	// +listMapKey=port
	Field []Item
}

// +k8s:openapi-gen=true
type Item struct {
	Protocol string
	Port     int
	// +optional
	A int `json:"a"`
	// +optional
	B int `json:"b,omitempty"`
	// +optional
	C int `json:"c,omitEmpty"`
}
