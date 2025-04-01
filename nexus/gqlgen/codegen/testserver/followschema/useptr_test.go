// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package followschema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserPtr(t *testing.T) {
	s := &Stub{}
	r := reflect.TypeOf(s.QueryResolver.OptionalUnion)
	require.True(t, r.Out(0).Kind() == reflect.Interface)
}
