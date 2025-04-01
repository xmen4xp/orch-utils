// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package errors

// Validation represents a failure of a precondition
type Validation struct {
	code    int32
	Name    string
	In      string
	Value   interface{}
	Valid   interface{}
	message string
	Values  []interface{}
}

func (e *Validation) Error() string {
	return e.message
}

// Code the error code
func (e *Validation) Code() int32 {
	return e.code
}

// ValidateName produces an error message name for an aliased property
func (e *Validation) ValidateName(name string) *Validation {
	if e.Name == "" && name != "" {
		e.Name = name
		e.message = name + e.message
	}
	return e
}
