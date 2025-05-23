// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractModuleName(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	modDir := filepath.Join(wd, "..", "..", "testdata")
	content, err := os.ReadFile(filepath.Join(modDir, "gomod-with-leading-comments.mod"))
	require.NoError(t, err)
	assert.Equal(t, "github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen", extractModuleName(content))
}

func TestImportPathForDir(t *testing.T) {
	wd, err := os.Getwd()

	require.NoError(t, err)

	assert.Equal(t, "github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/internal/code", ImportPathForDir(wd))

	assert.Equal(t, "github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/internal/code", ImportPathForDir(wd))
	assert.Equal(t, "github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/api", ImportPathForDir(filepath.Join(wd, "..", "..", "api")))

	// doesnt contain go code, but should still give a valid import path
	assert.Equal(t, "github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/docs", ImportPathForDir(filepath.Join(wd, "..", "..", "docs")))

	// directory does not exist
	assert.Equal(t, "github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/dos", ImportPathForDir(filepath.Join(wd, "..", "..", "dos")))

	// out of module
	assert.Equal(t, "", ImportPathForDir(filepath.Join(wd, "..", "..", "..")))

	if runtime.GOOS == "windows" {
		assert.Equal(t, "", ImportPathForDir("C:/doesnotexist"))
	} else {
		assert.Equal(t, "", ImportPathForDir("/doesnotexist"))
	}
}

func TestNameForDir(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	assert.Equal(t, "tmp", NameForDir("/tmp"))
	assert.Equal(t, "code", NameForDir(wd))
	assert.Equal(t, "docs", NameForDir(wd+"/../../docs"))
	assert.Equal(t, "main", NameForDir(wd+"/../.."))
}
