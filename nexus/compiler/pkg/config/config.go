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

	// ConversionWebhookServiceName/Namespace optionally pin the Service
	// fronting the conversion webhook (see nexus-api-gw's /convert handler)
	// that CRDs with more than one version (see pkg/parser/node_version.go)
	// point their spec.conversion.webhook.clientConfig at. When unset, the
	// compiler emits placeholder tokens instead (see
	// generator.ConversionWebhookServiceName{,Namespace}Placeholder) that
	// datamodel_installer.sh resolves at install time via nexus-api-gw's
	// stable `app.kubernetes.io/component=nexus-api-gw` label selector -
	// the Service's actual name is Helm-release-dependent, so it can't be
	// known at compile time in general.
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
