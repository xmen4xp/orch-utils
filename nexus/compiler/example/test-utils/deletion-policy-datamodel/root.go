// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// This datamodel is a parse-only fixture used by the compiler parser unit
// tests to exercise the nexus-deletion-policy annotation. It is intentionally
// kept out of example/datamodel so it is not part of the generated golden
// output.
package root

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

// nexus-deletion-policy: restrict
// nexus-deletion-restrict-children: AISlice
type Root struct {
	nexus.Node

	AISlice AISlice `nexus:"child"`
	Foo     Foo     `nexus:"child"`
}

// nexus-deletion-policy: cascade
type AISlice struct {
	nexus.Node
}

// Foo has no deletion policy annotation (defaults to cascade).
type Foo struct {
	nexus.Node
}
