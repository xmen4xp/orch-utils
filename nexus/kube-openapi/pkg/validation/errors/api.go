// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
)

// Error represents a error interface all swagger framework errors implement
type Error interface {
	error
	Code() int32
}

type apiError struct {
	code    int32
	message string
}

func (a *apiError) Error() string {
	return a.message
}

func (a *apiError) Code() int32 {
	return a.code
}

// New creates a new API error with a code and a message
func New(code int32, message string, args ...interface{}) Error {
	if len(args) > 0 {
		return &apiError{code, fmt.Sprintf(message, args...)}
	}
	return &apiError{code, message}
}
