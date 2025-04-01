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
package projectactivewatcher

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
)

type ActiveWatcherStatus string

const (
	StatusIndicationInProgress ActiveWatcherStatus = "STATUS_INDICATION_IN_PROGRESS"
	StatusIndicationError      ActiveWatcherStatus = "STATUS_INDICATION_ERROR"
	StatusIndicationIdle       ActiveWatcherStatus = "STATUS_INDICATION_IDLE"
)

type ProjectActiveWatcher struct {
	nexus.Node

	// StatusIndicator specifies the current status of the project (e.g., error, in progress, idle),
	// to acknowledge project create/delete notification.
	StatusIndicator ActiveWatcherStatus

	// Additional information or message about the error state of the project.
	Message string

	// Timestamp of when the status was last updated.
	TimeStamp uint64
}
