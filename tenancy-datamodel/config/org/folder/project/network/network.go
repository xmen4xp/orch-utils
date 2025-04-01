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

package network

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
)

// REST API to CRUD project.
var NetworkRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1/projects/{project.Project}/networks/{network.Network}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1/projects/{project.Project}/networks",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// TO-DO: Revisit nexus-deferred-delete after the implementation of the controller is further along and
//        we are able to implement delete logic.

// nexus-rest-api-gen:NetworkRestAPISpec
// nexus-deferred-delete: false
type Network struct {
	nexus.Node

	Type        NetworkType
	Description string `json:"description,omitempty"`

	// Status to be associated with this Project.
	Status NetworkStatus `nexus:"status"`
}

//nolint:revive // Per requirement.
type NetworkType string

//nolint:revive // Per requirement.
type NetworkStatus struct {
	CurrentState string
}

const (
	AppplicationMesh NetworkType = "application-mesh"
)
