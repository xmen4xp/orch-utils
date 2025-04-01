// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mage

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Namespace for building Go binaries
type Go mg.Namespace

const (
	binaryName   = "ktc"
	commitEnvVar = "KTC_GIT_COMMIT"
)

var (
	goEnvs = map[string]string{
		"GOPRIVATE": "github.com/open-edge-platform/*",
	}
)

func (Go) mod() error {
	return sh.RunWithV(goEnvs, "go", "mod", "tidy")
}

// Builds the Go Binaries without the release flags set
func (Go) Build() error {
	mg.Deps(Go.mod)
	if err := sh.RunV("mkdir", "-p", "./bin"); err != nil {
		return err
	}

	commit, err := getGitCommit()
	if err != nil {
		return err
	}

	ldFlags := "-ldflags=-w -s -buildid= -X 'main.Commit=" + commit + "'"
	return sh.RunWithV(goEnvs, "go", "build", ldFlags, "-trimpath", "-o", "./bin/"+binaryName, "./cmd/")
}

func getGitCommit() (string, error) {
	var buffer bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Stdout = &buffer
	if err := cmd.Run(); err != nil {
		// checking git has failed, check if there is an env var
		return getGitCommitEnvVar()
	}
	return strings.TrimSuffix(buffer.String(), "\n"), nil
}

func getGitCommitEnvVar() (string, error) {
	commitID := os.Getenv(commitEnvVar)
	if len(commitID) == 0 {
		return "", fmt.Errorf("no %s set in environment", commitEnvVar)
	}
	return commitID, nil
}
