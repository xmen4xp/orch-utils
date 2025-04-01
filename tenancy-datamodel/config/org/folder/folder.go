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

package folder

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/config/org/folder/project"
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
)

type Folder struct {
	nexus.Node

	// Projects associated with this Folder.
	Projects project.Project `nexus:"children"`
}
