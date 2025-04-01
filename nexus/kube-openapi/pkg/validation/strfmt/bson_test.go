// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package strfmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBSONObjectId_fullCycle(t *testing.T) {
	id := NewObjectId("507f1f77bcf86cd799439011")
	bytes, err := id.MarshalText()
	assert.NoError(t, err)

	var idCopy ObjectId

	err = idCopy.UnmarshalText(bytes)
	assert.NoError(t, err)
	assert.Equal(t, id, idCopy)

	jsonBytes, err := id.MarshalJSON()
	assert.NoError(t, err)

	err = idCopy.UnmarshalJSON(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, id, idCopy)
}

func TestDeepCopyObjectId(t *testing.T) {
	id := NewObjectId("507f1f77bcf86cd799439011")
	in := &id

	out := new(ObjectId)
	in.DeepCopyInto(out)
	assert.Equal(t, in, out)

	out2 := in.DeepCopy()
	assert.Equal(t, in, out2)

	var inNil *ObjectId
	out3 := inNil.DeepCopy()
	assert.Nil(t, out3)
}
