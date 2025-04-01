// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

// SchemaValidatorOptions defines optional rules for schema validation
type SchemaValidatorOptions struct {
	validationRulesEnabled bool
}

// Option sets optional rules for schema validation
type Option func(*SchemaValidatorOptions)

// Options returns current options
func (svo SchemaValidatorOptions) Options() []Option {
	return []Option{}
}
