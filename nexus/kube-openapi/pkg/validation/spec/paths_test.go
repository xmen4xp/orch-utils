// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var paths = Paths{
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{"x-framework": "go-swagger"}},
	Paths: map[string]PathItem{
		"/": {
			Refable: Refable{Ref: MustCreateRef("cats")},
		},
	},
}

const pathsJSON = `{"x-framework":"go-swagger","/":{"$ref":"cats"}}`

func TestIntegrationPaths(t *testing.T) {
	var actual Paths
	if assert.NoError(t, json.Unmarshal([]byte(pathsJSON), &actual)) {
		assert.EqualValues(t, actual, paths)
	}

	assertParsesJSON(t, pathsJSON, paths)

}
