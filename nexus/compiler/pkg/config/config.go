// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"io"
	"os"

	"github.com/ghodss/yaml"
)

var ConfigInstance *Config

type Config struct {
	GroupName               string   `yaml:"groupName"`
	CrdModulePath           string   `yaml:"crdModulePath"`
	IgnoredDirs             []string `yaml:"ignoredDirs"`
	IgnoredParentPathParams []string `yaml:"ignoredParentPathParams"`

	// ConversionWebhookServiceName/Namespace identify the Service fronting
	// the conversion webhook (see nexus-api-gw's /convert handler) that CRDs
	// with more than one version (see pkg/parser/node_version.go) point
	// their spec.conversion.webhook.clientConfig at. Defaults to
	// "nexus-api-gw" in whatever namespace the datamodel is installed into
	// when unset.
	ConversionWebhookServiceName      string `yaml:"conversionWebhookServiceName"`
	ConversionWebhookServiceNamespace string `yaml:"conversionWebhookServiceNamespace"`
}

func LoadConfig(configFile string) (*Config, error) {
	var config *Config
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %s", err)
	}
	configStr, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}
	err = yaml.Unmarshal(configStr, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %s", err)
	}
	return config, nil
}
