// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package codegen

func (o *Object) UniqueFields() map[string][]*Field {
	m := map[string][]*Field{}

	for _, f := range o.Fields {
		m[f.GoFieldName] = append(m[f.GoFieldName], f)
	}

	return m
}
