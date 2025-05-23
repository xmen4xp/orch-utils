// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func float64Ptr(f float64) *float64 {
	return &f
}
func int64Ptr(f int64) *int64 {
	return &f
}

var header = Header{
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{
		"x-framework": "swagger-go",
	}},
	HeaderProps: HeaderProps{Description: "the description of this header"},
	SimpleSchema: SimpleSchema{
		Items: &Items{
			Refable: Refable{Ref: MustCreateRef("Cat")},
		},
		Type:    "string",
		Format:  "date",
		Default: "8",
	},
	CommonValidations: CommonValidations{
		Maximum:          float64Ptr(100),
		ExclusiveMaximum: true,
		ExclusiveMinimum: true,
		Minimum:          float64Ptr(5),
		MaxLength:        int64Ptr(100),
		MinLength:        int64Ptr(5),
		Pattern:          "\\w{1,5}\\w+",
		MaxItems:         int64Ptr(100),
		MinItems:         int64Ptr(5),
		UniqueItems:      true,
		MultipleOf:       float64Ptr(5),
		Enum:             []interface{}{"hello", "world"},
	},
}

const headerJSON = `{
  "items": {
    "$ref": "Cat"
  },
  "x-framework": "swagger-go",
  "description": "the description of this header",
  "maximum": 100,
  "minimum": 5,
  "exclusiveMaximum": true,
  "exclusiveMinimum": true,
  "maxLength": 100,
  "minLength": 5,
  "pattern": "\\w{1,5}\\w+",
  "maxItems": 100,
  "minItems": 5,
  "uniqueItems": true,
  "multipleOf": 5,
  "enum": ["hello", "world"],
  "type": "string",
  "format": "date",
  "default": "8"
}`

// cmp.Diff panics when reflecting unexported fields under jsonreference.Ref
// a custom comparator is required
var swaggerDiffOptions = []cmp.Option{cmp.Comparer(func(a Ref, b Ref) bool {
	return a.String() == b.String()
})}

func TestIntegrationHeader(t *testing.T) {
	var actual Header
	if assert.NoError(t, json.Unmarshal([]byte(headerJSON), &actual)) {
		if !reflect.DeepEqual(header, actual) {
			t.Fatal(cmp.Diff(header, actual, swaggerDiffOptions...))
		}
	}

	assertParsesJSON(t, headerJSON, header)
}

// Makes sure that a Header unmarshaled from known good JSON, and one unmarshaled
// from generated JSON are equivalent.
func TestHeaderSerialization(t *testing.T) {
	generatedJSON, err := json.Marshal(header)
	require.NoError(t, err)

	generatedJSONActual := Header{}
	require.NoError(t, json.Unmarshal(generatedJSON, &generatedJSONActual))
	if !reflect.DeepEqual(header, generatedJSONActual) {
		t.Fatal(cmp.Diff(header, generatedJSONActual, swaggerDiffOptions...))
	}

	goodJSONActual := Header{}
	require.NoError(t, json.Unmarshal([]byte(headerJSON), &goodJSONActual))
	if !reflect.DeepEqual(header, goodJSONActual) {
		t.Fatal(cmp.Diff(header, goodJSONActual, swaggerDiffOptions...))
	}
}
