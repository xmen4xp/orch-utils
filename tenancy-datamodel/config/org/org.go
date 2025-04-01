/*
 * Copyright (C) 2025 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions
 * and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/config/org/folder"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
)

type TenancyRequestStatus string

const (
	StatusIndicationUnspecified TenancyRequestStatus = "STATUS_INDICATION_UNSPECIFIED"
	StatusIndicationError       TenancyRequestStatus = "STATUS_INDICATION_ERROR"
	StatusIndicationInProgress  TenancyRequestStatus = "STATUS_INDICATION_IN_PROGRESS"
	StatusIndicationIdle        TenancyRequestStatus = "STATUS_INDICATION_IDLE"
)

// REST API to CRUD org.
var OrgRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1/orgs/{org.Org}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1/orgs",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:OrgRestAPISpec
// nexus-deferred-delete: true
type Org struct {
	nexus.Node

	// Description of org.
	Description string

	// Folders associated with this org.
	Folders folder.Folder `nexus:"children"`

	// Status associated with this org.
	OrgStatus OrgStatus `nexus:"status"`
}

//nolint:revive // Per requirement.
type OrgStatus struct {
	// StatusIndicator specifies the current status of the org (e.g., error, in progress, idle).
	StatusIndicator TenancyRequestStatus

	// Additional information or message about the error state of the org.
	Message string

	// Timestamp of when the status was last updated.
	TimeStamp uint64

	// Unigue ID assigned to this org.
	UID string
}
