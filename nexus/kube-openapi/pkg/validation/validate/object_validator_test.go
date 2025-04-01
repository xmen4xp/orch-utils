// Copyright 2017 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func itemsFixture() map[string]interface{} {
	return map[string]interface{}{
		"type":  "array",
		"items": "dummy",
	}
}

func expectAllValid(t *testing.T, ov valueValidator, dataValid, dataInvalid map[string]interface{}) {
	res := ov.Validate(dataValid)
	assert.Equal(t, 0, len(res.Errors))

	res = ov.Validate(dataInvalid)
	assert.Equal(t, 0, len(res.Errors))
}

func expectOnlyInvalid(t *testing.T, ov valueValidator, dataValid, dataInvalid map[string]interface{}) {
	res := ov.Validate(dataValid)
	assert.Equal(t, 0, len(res.Errors))

	res = ov.Validate(dataInvalid)
	assert.NotEqual(t, 0, len(res.Errors))
}

func TestItemsMustBeTypeArray(t *testing.T) {
	ov := new(objectValidator)
	dataValid := itemsFixture()
	dataInvalid := map[string]interface{}{
		"type":  "object",
		"items": "dummy",
	}
	expectAllValid(t, ov, dataValid, dataInvalid)
}

func TestItemsMustHaveType(t *testing.T) {
	ov := new(objectValidator)
	dataValid := itemsFixture()
	dataInvalid := map[string]interface{}{
		"items": "dummy",
	}
	expectAllValid(t, ov, dataValid, dataInvalid)
}

func TestTypeArrayMustHaveItems(t *testing.T) {
	ov := new(objectValidator)
	dataValid := itemsFixture()
	dataInvalid := map[string]interface{}{
		"type": "array",
		"key":  "dummy",
	}
	expectAllValid(t, ov, dataValid, dataInvalid)
}

// Test edge cases in object_validator which are difficult
// to simulate with specs
// (this one is a trivial, just to check all methods are filled)
func TestObjectValidator_EdgeCases(t *testing.T) {
	s := objectValidator{}
	s.SetPath("path")
	assert.Equal(t, "path", s.Path)
}
