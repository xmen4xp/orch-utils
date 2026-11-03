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

// Org is the default/storage version. No version annotations needed - this
// is exactly how every pre-existing (unversioned) DSL type looks.
type Org struct {
	nexus.Node
	Description string
}

// nexus-version: v2
// nexus-kind: Org
type OrgV2 struct {
	nexus.Node
	Description string
	DisplayName string
}

// conversionHub is a stand-in for sigs.k8s.io/controller-runtime's
// conversion.Hub in this fixture, since the compiler only checks for the
// ConvertTo/ConvertFrom method names and does not type-check their signature.
type conversionHub interface{}

func (o *OrgV2) ConvertTo(hub conversionHub) error   { return nil }
func (o *OrgV2) ConvertFrom(hub conversionHub) error { return nil }
