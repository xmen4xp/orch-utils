// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package otherpkg

type (
	Scalar string
	Map    map[string]string
	Slice  []string
)

type Struct struct {
	Name Scalar
	Desc *Scalar
}
