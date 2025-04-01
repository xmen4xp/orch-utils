// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package root

import (
	global "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/test-utils/global-package/project"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Project global.Project `nexus:"child"`
	IsRoot  IsRoot         // <--- to verify alias type
}

type IsRoot bool
