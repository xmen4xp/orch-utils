// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/test-utils/non-singleton-root/config"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Project struct {
	nexus.SingletonNode
	Key    string
	Config config.Config `nexus:"child"`
}
