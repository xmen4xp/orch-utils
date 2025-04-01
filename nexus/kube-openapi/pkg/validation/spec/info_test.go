// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const infoJSON = `{
	"description": "A sample API that uses a petstore as an example to demonstrate features in ` +
	`the swagger-2.0 specification",
	"title": "Swagger Sample API",
	"termsOfService": "http://helloreverb.com/terms/",
	"contact": {
		"name": "wordnik api team",
		"url": "http://developer.wordnik.com"
	},
	"license": {
		"name": "Creative Commons 4.0 International",
		"url": "http://creativecommons.org/licenses/by/4.0/"
	},
	"version": "1.0.9-abcd",
	"x-framework": "go-swagger"
}`

var info = Info{
	InfoProps: InfoProps{
		Version: "1.0.9-abcd",
		Title:   "Swagger Sample API",
		Description: "A sample API that uses a petstore as an example to demonstrate features in " +
			"the swagger-2.0 specification",
		TermsOfService: "http://helloreverb.com/terms/",
		Contact:        &ContactInfo{Name: "wordnik api team", URL: "http://developer.wordnik.com"},
		License: &License{
			Name: "Creative Commons 4.0 International",
			URL:  "http://creativecommons.org/licenses/by/4.0/",
		},
	},
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{"x-framework": "go-swagger"}},
}

func TestIntegrationInfo_Serialize(t *testing.T) {
	b, err := json.MarshalIndent(info, "", "\t")
	if assert.NoError(t, err) {
		assert.Equal(t, infoJSON, string(b))
	}
}

func TestIntegrationInfo_Deserialize(t *testing.T) {
	actual := Info{}
	err := json.Unmarshal([]byte(infoJSON), &actual)
	if assert.NoError(t, err) {
		assert.EqualValues(t, info, actual)
	}
}
