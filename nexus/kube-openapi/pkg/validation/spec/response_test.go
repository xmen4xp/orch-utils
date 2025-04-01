// Copyright 2017 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var response = Response{
	Refable: Refable{Ref: MustCreateRef("Dog")},
	VendorExtensible: VendorExtensible{
		Extensions: map[string]interface{}{
			"x-go-name": "PutDogExists",
		},
	},
	ResponseProps: ResponseProps{
		Description: "Dog exists",
		Schema:      &Schema{SchemaProps: SchemaProps{Type: []string{"string"}}},
	},
}

const responseJSON = `{
	"$ref": "Dog",
	"x-go-name": "PutDogExists",
	"description": "Dog exists",
	"schema": {
		"type": "string"
	}
}`

func TestIntegrationResponse(t *testing.T) {
	var actual Response
	if assert.NoError(t, json.Unmarshal([]byte(responseJSON), &actual)) {
		assert.EqualValues(t, actual, response)
	}

	assertParsesJSON(t, responseJSON, response)
}
