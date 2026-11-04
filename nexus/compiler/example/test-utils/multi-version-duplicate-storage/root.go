// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package root

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Org Org `nexus:"children"`
}

type Org struct {
	nexus.Node
	Description string
}

// nexus-version: v2
// nexus-kind: Org
// nexus-storage-version: true
//
// Declaring a second storage version for the same Kind is invalid - exactly
// one version of a Kind may be the storage version.
type OrgV2 struct {
	nexus.Node
	Description string
	DisplayName string
}
