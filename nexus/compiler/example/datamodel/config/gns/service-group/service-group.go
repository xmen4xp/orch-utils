// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package servicegroup

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

// nexus-secret-spec:ApiKeySecretSpec
type SvcGroup struct {
	nexus.Node
	DisplayName string
	Description string
	Color       string
}

type SvcGroupLinkInfo struct {
	nexus.Node
	ClusterName string
	DomainName  string
	ServiceName string
	ServiceType string
}
