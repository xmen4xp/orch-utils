// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRootFieldContext(t *testing.T) {
	require.Nil(t, GetRootFieldContext(context.Background()))

	rc := &RootFieldContext{}
	require.Equal(t, rc, GetRootFieldContext(WithRootFieldContext(context.Background(), rc)))
}
