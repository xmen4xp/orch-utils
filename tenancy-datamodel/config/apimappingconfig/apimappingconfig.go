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

package apimappingconfig

import (
	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/nexus/base/nexus"
)

//nolint:tagliatelle // Per requirement.
type APIMappingConfig struct {
	nexus.Node

	SpecGenEnabled bool      `json:"specGenEnabled"`
	RepoConf       RepoConf  `json:"repoConf"`
	Mappings       []Mapping `json:"mappings"`
	Backend        Backend   `json:"backend"`
}

//nolint:tagliatelle // Per requirement.
type RepoConf struct {
	URL          string `json:"url"`
	Tag          string `json:"tag"`
	SpecFilePath string `json:"specFilePath"`
}

//nolint:tagliatelle // Per requirement.
type Mapping struct {
	ExternalURI string `json:"externalURI"`
	ServiceURI  string `json:"serviceURI"`
}

type Backend struct {
	Service string `json:"service"`
	Port    uint32 `json:"port"`
}
