// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mage

import (
	"github.com/magefile/mage/sh"
)

func (Lint) yaml() error {
	return sh.RunV(
		"yamllint",
		"-c",
		"tools/yamllint-conf.yaml",
		".",
	)
}

// Formats the Go code
func (Lint) Gofmt() error {
	return sh.RunV(
		"go",
		"fmt",
		"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/...",
	)
}
