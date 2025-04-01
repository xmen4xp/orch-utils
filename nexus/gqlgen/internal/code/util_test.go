// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeVendor(t *testing.T) {
	require.Equal(t, "bar/baz", NormalizeVendor("foo/vendor/bar/baz"))
	require.Equal(t, "[]bar/baz", NormalizeVendor("[]foo/vendor/bar/baz"))
	require.Equal(t, "*bar/baz", NormalizeVendor("*foo/vendor/bar/baz"))
	require.Equal(t, "*[]*bar/baz", NormalizeVendor("*[]*foo/vendor/bar/baz"))
	require.Equal(t, "[]*bar/baz", NormalizeVendor("[]*foo/vendor/bar/baz"))
}
