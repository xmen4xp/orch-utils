// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean(t *testing.T) {
	assert.Equal(t, "true", doBooleanMarshal(true))
	assert.Equal(t, "false", doBooleanMarshal(false))
}

func doBooleanMarshal(b bool) string {
	var buf bytes.Buffer
	MarshalBoolean(b).MarshalGQL(&buf)
	return buf.String()
}
