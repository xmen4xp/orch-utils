// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test edge cases in slice_validator which are difficult
// to simulate with specs
// (this one is a trivial, just to check all methods are filled)
func TestSliceValidator_EdgeCases(t *testing.T) {
	s := schemaSliceValidator{}
	s.SetPath("path")
	assert.Equal(t, "path", s.Path)

	r := s.Validate(nil)
	assert.NotNil(t, r)
	assert.True(t, r.IsValid())
}
