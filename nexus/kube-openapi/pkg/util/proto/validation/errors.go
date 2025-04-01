// Copyright 2017 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
)

type errors struct {
	errors []error
}

func (e *errors) Errors() []error {
	return e.errors
}

func (e *errors) AppendErrors(err ...error) {
	e.errors = append(e.errors, err...)
}

type ValidationError struct {
	Path string
	Err  error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("ValidationError(%s): %v", e.Path, e.Err)
}

type InvalidTypeError struct {
	Path     string
	Expected string
	Actual   string
}

func (e InvalidTypeError) Error() string {
	return fmt.Sprintf("invalid type for %s: got %q, expected %q", e.Path, e.Actual, e.Expected)
}

type MissingRequiredFieldError struct {
	Path  string
	Field string
}

func (e MissingRequiredFieldError) Error() string {
	return fmt.Sprintf("missing required field %q in %s", e.Field, e.Path)
}

type UnknownFieldError struct {
	Path  string
	Field string
}

func (e UnknownFieldError) Error() string {
	return fmt.Sprintf("unknown field %q in %s", e.Field, e.Path)
}

type InvalidObjectTypeError struct {
	Path string
	Type string
}

func (e InvalidObjectTypeError) Error() string {
	return fmt.Sprintf("unknown object type %q in %s", e.Type, e.Path)
}
