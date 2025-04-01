// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package root

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/test-utils/invalid-type-name-datamodel/config"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Config config.Config `nexus:"child"`
}
