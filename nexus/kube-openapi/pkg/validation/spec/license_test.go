// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import "testing"

func TestIntegrationLicense(t *testing.T) {
	license := License{Name: "the name", URL: "the url"}
	const licenseJSON = `{"name":"the name","url":"the url"}`
	const licenseYAML = "name: the name\nurl: the url\n"

	assertSerializeJSON(t, license, licenseJSON)
	assertParsesJSON(t, licenseJSON, license)
}
