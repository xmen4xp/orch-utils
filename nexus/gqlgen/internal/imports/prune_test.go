// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package imports

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/internal/code"
)

func TestPrune(t *testing.T) {
	// prime the packages cache so that it's not considered uninitialized

	b, err := Prune("testdata/unused.go", mustReadFile("testdata/unused.go"), &code.Packages{})
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(string(mustReadFile("testdata/unused.expected.go")), "\r\n", "\n"), string(b))
}

func mustReadFile(filename string) []byte {
	b, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return b
}
