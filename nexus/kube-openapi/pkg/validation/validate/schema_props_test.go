// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test edge cases in schema_props_validator which are difficult
// to simulate with specs
// (this one is a trivial, just to check all methods are filled)
func TestSchemaPropsValidator_EdgeCases(t *testing.T) {
	s := schemaPropsValidator{}
	s.SetPath("path")
	assert.Equal(t, "path", s.Path)
}
