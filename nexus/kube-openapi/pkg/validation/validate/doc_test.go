// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate_test

import (
	"encoding/json"
	"fmt"
	"testing"

	// Spec loading
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/strfmt"   // OpenAPI format extensions
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/validate" // This package
)

func ExampleAgainstSchema() {
	// Example using encoding/json as unmarshaller
	var schemaJSON = `
{
    "properties": {
        "name": {
            "type": "string",
            "pattern": "^[A-Za-z]+$",
            "minLength": 1
        }
	},
    "patternProperties": {
	  "address-[0-9]+": {
         "type": "string",
         "pattern": "^[\\s|a-z]+$"
	  }
    },
    "required": [
        "name"
    ],
	"additionalProperties": false
}`

	schema := new(spec.Schema)
	_ = json.Unmarshal([]byte(schemaJSON), schema)

	input := map[string]interface{}{}

	// JSON data to validate
	inputJSON := `{"name": "Ivan","address-1": "sesame street"}`
	_ = json.Unmarshal([]byte(inputJSON), &input)

	// strfmt.Default is the registry of recognized formats
	err := validate.AgainstSchema(schema, input, strfmt.Default)
	if err != nil {
		fmt.Printf("JSON does not validate against schema: %v", err)
	} else {
		fmt.Printf("OK")
	}
	// Output:
	// OK
}

func TestValidate_Issue112(t *testing.T) {
	t.Run("returns no error on body includes `items` key", func(t *testing.T) {
		body := map[string]interface{}{"items1": nil}
		err := validate.AgainstSchema(getSimpleSchema(), body, strfmt.Default)
		require.NoError(t, err)
	})

	t.Run("returns no error when body includes `items` key", func(t *testing.T) {
		body := map[string]interface{}{"items": nil}
		err := validate.AgainstSchema(getSimpleSchema(), body, strfmt.Default)
		require.NoError(t, err)
	})
}

func getSimpleSchema() *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: spec.StringOrArray{"object"},
		},
	}
}
