// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mage

import (
	"github.com/magefile/mage/mg"
)

type Lint mg.Namespace

// Lint everything.
func (l Lint) All() error {
	if err := l.yaml(); err != nil {
		return err
	}
	return nil
}

// Namespace for container builds that weren't already migrated to Mage
// These may be added to containers.go
type Binary mg.Namespace

// Mage function that just calls other functions for building the container without the release flags
func (Binary) Build() {
	mg.Deps(Go.Build)
}
