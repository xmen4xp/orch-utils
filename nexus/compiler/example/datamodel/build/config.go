// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

// BuildTestStruct struct is in build directory and should be ignored by parser
type BuildTestStruct struct {
	nexus.Node
	SomeInt int
}
