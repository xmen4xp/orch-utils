// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	logMutex = &sync.Mutex{}
)

func TestDebug(t *testing.T) {
	tmpFile, _ := ioutil.TempFile("", "debug-test")
	tmpName := tmpFile.Name()
	defer func() {
		Debug = false
		// mutex for -race
		logMutex.Unlock()
		os.Remove(tmpName)
	}()

	// mutex for -race
	logMutex.Lock()
	Debug = true
	debugOptions()
	defer func() {
		validateLogger.SetOutput(os.Stdout)
	}()

	validateLogger.SetOutput(tmpFile)

	debugLog("A debug")
	Debug = false
	tmpFile.Close()

	flushed, _ := os.Open(tmpName)
	buf := make([]byte, 500)
	_, _ = flushed.Read(buf)
	validateLogger.SetOutput(os.Stdout)
	assert.Contains(t, string(buf), "A debug")
}
