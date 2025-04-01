//go:build mage

// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	// mage:import
	. "github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/mage" //nolint: revive
)

// To silence compiler's unused import error.
var _ = Lint{}
