// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package root

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/common-library/pkg/nexus"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/test-utils/group-name-with-hyphen-datamodel/project"
)

type Root struct {
	nexus.Node
	SomeRootData string
	Project      project.Project `nexus:"child"`
}
