// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package global

import (
	global "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/test-utils/global-package/config"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Project struct {
	nexus.SingletonNode
	Key    string
	Config global.Config `nexus:"child"`
}
