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

package config

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/config/apimappingconfig"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/config/org"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/config/orgwatcher"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/config/projectwatcher"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
)

// Config tree.
type Config struct {
	nexus.SingletonNode

	// Organizations created by User.
	Orgs org.Org `nexus:"children"`

	// Components to be notified of org create/delete.
	OrgWatchers orgwatcher.OrgWatcher `nexus:"children"`

	// APIMappings to support backend services.
	APIMappings apimappingconfig.APIMappingConfig `nexus:"children"`

	// Components to be notified of project create/delete.
	ProjectWatchers projectwatcher.ProjectWatcher `nexus:"children"`
}
