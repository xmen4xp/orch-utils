// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

type Config struct {
	nexus.Node
	Id          string
	FooResource Resource
}

type Resource struct {
	Name string
}
