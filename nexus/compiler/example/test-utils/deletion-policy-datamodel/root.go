// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// This datamodel is a parse-only fixture used by the compiler parser unit
// tests to exercise the nexus-on-delete child tag. It is intentionally kept
// out of example/datamodel so it is not part of the generated golden output.
package root

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node

	// AISlice blocks deletion of Root while it exists.
	AISlice AISlice `nexus:"child" nexus-on-delete:"restrict"`
	// Foo has no on-delete tag; it defaults to cascade.
	Foo Foo `nexus:"child"`
}

type AISlice struct {
	nexus.Node
}

type Foo struct {
	nexus.Node
}
