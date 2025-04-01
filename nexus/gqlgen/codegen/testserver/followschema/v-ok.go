// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package followschema

// VOkCaseValue model
type VOkCaseValue struct{}

func (v VOkCaseValue) Value() (string, bool) {
	return "hi", true
}

// VOkCaseNil model
type VOkCaseNil struct{}

func (v VOkCaseNil) Value() (string, bool) {
	return "", false
}
