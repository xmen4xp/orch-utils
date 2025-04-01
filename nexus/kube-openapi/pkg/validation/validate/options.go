// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

import "sync"

// Opts specifies validation options for a SpecValidator.
//
// NOTE: other options might be needed, for example a go-swagger specific mode.
type Opts struct {
	ContinueOnErrors bool // true: continue reporting errors, even if spec is invalid
}

var (
	defaultOpts      = Opts{ContinueOnErrors: false} // default is to stop validation on errors
	defaultOptsMutex = &sync.Mutex{}
)

// SetContinueOnErrors sets global default behavior regarding spec validation errors reporting.
//
// For extended error reporting, you most likely want to set it to true.
// For faster validation, it's better to give up early when a spec is detected as invalid: set it to false (this is the default).
//
// Setting this mode does NOT affect the validation status.
//
// NOTE: this method affects global defaults. It is not suitable for a concurrent usage.
func SetContinueOnErrors(c bool) {
	defer defaultOptsMutex.Unlock()
	defaultOptsMutex.Lock()
	defaultOpts.ContinueOnErrors = c
}
