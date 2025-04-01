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

package runtimeproject

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/runtime/org/folder/project/projectactivewatcher"
)

type RuntimeProject struct {
	nexus.Node

	// Indicates that project has been deleted by the User.
	Deleted bool

	// Watchers actively watching this project for create, delete.
	ActiveWatchers projectactivewatcher.ProjectActiveWatcher `nexus:"children"`
}
