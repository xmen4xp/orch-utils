// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var parameter = Parameter{
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{
		"x-framework": "swagger-go",
	}},
	Refable: Refable{Ref: MustCreateRef("Dog")},
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
	SimpleSchema: SimpleSchema{
		Type:             "string",
		Format:           "date",
		CollectionFormat: "csv",
		Items: &Items{
			Refable: Refable{Ref: MustCreateRef("Cat")},
		},
		Default: "8",
	},
	ParamProps: ParamProps{
		Name:        "param-name",
		In:          "header",
		Required:    true,
		Schema:      &Schema{SchemaProps: SchemaProps{Type: []string{"string"}}},
		Description: "the description of this parameter",
	},
}

var parameterJSON = `{
	"items": {
		"$ref": "Cat"
	},
	"x-framework": "swagger-go",
  "$ref": "Dog",
  "description": "the description of this parameter",
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
	"name": "param-name",
	"in": "header",
	"required": true,
	"schema": {
		"type": "string"
	},
	"collectionFormat": "csv",
	"default": "8"
}`

func TestIntegrationParameter(t *testing.T) {
	var actual Parameter
	if assert.NoError(t, json.Unmarshal([]byte(parameterJSON), &actual)) {
		assert.EqualValues(t, actual, parameter)
	}

	assertParsesJSON(t, parameterJSON, parameter)
}
