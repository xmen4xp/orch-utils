// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// pin pointing go-swagger/go-swagger#1816 issue with cloning ref's
func TestCloneRef(t *testing.T) {
	var b bytes.Buffer
	src := MustCreateRef("#/definitions/test")
	err := gob.NewEncoder(&b).Encode(&src)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	jazon, err := json.Marshal(src)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, `{"$ref":"#/definitions/test"}`, string(jazon))
}
