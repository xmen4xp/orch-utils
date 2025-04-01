// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	v := &Validation{Name: "myValidation", message: "myMessage"}

	// unchanged
	vv := v.ValidateName("")
	assert.EqualValues(t, "myValidation", vv.Name)
	assert.EqualValues(t, "myMessage", vv.message)

	// unchanged
	vv = v.ValidateName("myNewName")
	assert.EqualValues(t, "myValidation", vv.Name)
	assert.EqualValues(t, "myMessage", vv.message)

	v.Name = ""

	// unchanged
	vv = v.ValidateName("")
	assert.EqualValues(t, "", vv.Name)
	assert.EqualValues(t, "myMessage", vv.message)

	// forced
	vv = v.ValidateName("myNewName")
	assert.EqualValues(t, "myNewName", vv.Name)
	assert.EqualValues(t, "myNewNamemyMessage", vv.message)
}
