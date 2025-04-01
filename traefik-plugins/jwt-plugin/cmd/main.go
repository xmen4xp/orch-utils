// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	_ "github.com/team-carepay/traefik-jwt-plugin"
)

func main() {
	// This file is created simply to vendor the package in the import statement.
	// After downloading files to the vendor folder, the files matching wildcard package*json
	// were removed to reduce the size of the auto-generated config map.
}
