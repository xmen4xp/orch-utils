// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

func (Lint) helm() error {
	charts, err := listCharts()
	if err != nil {
		return err
	}
	for _, chart := range charts {
		fmt.Println(chart)
		if err := sh.RunV("helm", "lint", chart); err != nil {
			return err
		}
	}
	return nil
}

func (Lint) yaml() error {
	return sh.RunV("yamllint", "-c", "tools/yamllint-conf.yaml", ".")
}

func listCharts() ([]string, error) {
	var charts []string
	err := filepath.Walk("charts", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.Count(path, string(os.PathSeparator)) == 1 {
			charts = append(charts, path)
		}
		return nil
	})
	return charts, err
}

// Lint golang files.
func (Lint) golang() error {
	_ = sh.RunV("golangci-lint", "--version")
	return sh.RunV("golangci-lint", "run", "-v", "--timeout", "5m0s")
}

func (Lint) dockerfiles() error {
	var lintErrors []error
	dockerfiles, err := listDockerfiles()
	if err != nil {
		return err
	}
	for _, dockerfile := range dockerfiles {
		fmt.Println(dockerfile)
		if err := sh.RunV("hadolint", dockerfile); err != nil {
			lintErrors = append(lintErrors, err)
		}
	}
	if len(lintErrors) > 0 {
		return lintErrors[0]
	}
	return nil
}

func listDockerfiles() ([]string, error) {
	var dockerfiles []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == "Dockerfile" {
			fmt.Println(path)
			dockerfiles = append(dockerfiles, path)
		}
		return nil
	})
	return dockerfiles, err
}
