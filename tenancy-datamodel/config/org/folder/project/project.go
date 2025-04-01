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

package project

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/config/org/folder/project/network"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
)

type TenancyRequestStatus string

const (
	StatusIndicationUnspecified TenancyRequestStatus = "STATUS_INDICATION_UNSPECIFIED"
	StatusIndicationError       TenancyRequestStatus = "STATUS_INDICATION_ERROR"
	StatusIndicationInProgress  TenancyRequestStatus = "STATUS_INDICATION_IN_PROGRESS"
	StatusIndicationIdle        TenancyRequestStatus = "STATUS_INDICATION_IDLE"
)

// REST API to CRUD project.
var ProjectRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1/projects/{project.Project}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1/projects",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:ProjectRestAPISpec
// nexus-deferred-delete: true
type Project struct {
	nexus.Node

	// Description of project.
	Description string

	// Networks associated with this org.
	Networks network.Network `nexus:"children"`

	// Status to be associated with this Project.
	ProjectStatus ProjectStatus `nexus:"status"`
}

//nolint:revive // Per requirement.
type ProjectStatus struct {
	// StatusIndicator specifies the current status of the project (e.g., error, in progress, idle).
	StatusIndicator TenancyRequestStatus

	// Additional information or message about the error state of the project.
	Message string

	// Timestamp of when the status was last updated.
	TimeStamp uint64

	// Unigue ID assigned to this project.
	UID string
}
