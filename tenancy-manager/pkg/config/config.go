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
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const defaultTimeout int32 = 300

type Config struct {
	OrgCreateTimeoutInSecs     int32 `yaml:"OrgCreateTimeoutInSecs"`
	OrgDeleteTimeoutInSecs     int32 `yaml:"OrgDeleteTimeoutInSecs"`
	ProjectCreateTimeoutInSecs int32 `yaml:"ProjectCreateTimeoutInSecs"`
	ProjectDeleteTimeoutInSecs int32 `yaml:"ProjectDeleteTimeoutInSecs"`
}

// LoadConfig loads configuration from a YAML file mounted in the specified path.
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config *Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return config, nil
}

// LoadConfig loads configuration from a YAML file mounted in the specified path.
func GetDefaultConfig() *Config {
	return &Config{
		OrgCreateTimeoutInSecs:     defaultTimeout,
		OrgDeleteTimeoutInSecs:     defaultTimeout,
		ProjectCreateTimeoutInSecs: defaultTimeout,
		ProjectDeleteTimeoutInSecs: defaultTimeout,
	}
}
