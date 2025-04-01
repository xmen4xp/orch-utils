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

package runtimeorg

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
	runtimefolder "github.com/open-edge-platform/orch-utils/tenancy-datamodel/runtime/org/folder"
	orgactivewatcher "github.com/open-edge-platform/orch-utils/tenancy-datamodel/runtime/org/orgactivewatcher"
)

type RuntimeOrg struct {
	nexus.Node

	// Indicates that org has been deleted by the User.
	Deleted bool

	// Projects associated with this org.
	Folders runtimefolder.RuntimeFolder `nexus:"children"`

	// Watchers actively watching this org for create, delete.
	ActiveWatchers orgactivewatcher.OrgActiveWatcher `nexus:"children"`
}
