// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"
)

func TestIntegrationExternalDocs(t *testing.T) {
	var extDocs = ExternalDocumentation{Description: "the name", URL: "the url"}
	const extDocsJSON = `{"description":"the name","url":"the url"}`
	assertSerializeJSON(t, extDocs, extDocsJSON)
	assertParsesJSON(t, extDocsJSON, extDocs)
}
