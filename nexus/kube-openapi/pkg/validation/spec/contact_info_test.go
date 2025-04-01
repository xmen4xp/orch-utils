// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"
)

const contactInfoJSON = `{"name":"wordnik api team","url":"http://developer.wordnik.com","email":"some@mailayada.dkdkd"}`
const contactInfoYAML = `name: wordnik api team
url: http://developer.wordnik.com
email: some@mailayada.dkdkd
`

var contactInfo = ContactInfo{
	Name:  "wordnik api team",
	URL:   "http://developer.wordnik.com",
	Email: "some@mailayada.dkdkd",
}

func TestIntegrationContactInfo(t *testing.T) {
	assertSerializeJSON(t, contactInfo, contactInfoJSON)
	assertParsesJSON(t, contactInfoJSON, contactInfo)
}
