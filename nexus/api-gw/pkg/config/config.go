// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

const APIGwConfigFileDefaullt = "/config/api-gw-config"

type Config struct {
	Server                  ServerConfig      `json:"server" yaml:"server"`
	EnableNexusRuntime      bool              `json:"enableNexusRuntime" yaml:"enableNexusRuntime,omitempty"`
	DisableAuthz            bool              `json:"disableAuthz" yaml:"disableAuthz,omitempty"`
	BackendService          string            `json:"backendService" yaml:"backendService,omitempty"`
	TenancyService          bool              `json:"tenancyService" yaml:"tenancyService,omitempty"`
	TenantAPIGwDomain       string            `json:"tenantApiGwDomain" yaml:"tenantApiGwDomain,omitempty"`
	CustomNotFoundPage      string            `json:"customNotFoundPage" yaml:"customNotFoundPage,omitempty"`
	ProxyInsecureSkipVerify bool              `json:"proxyInsecureSkipVerify" yaml:"proxyInsecureSkipVerify,omitempty"`
	HeaderAliases           map[string]string `json:"headerAliases" yaml:"headerAliases,omitempty"`
}

type ServerConfig struct {
	HTTPPort            string `json:"httpPort" yaml:"httpPort"`
	HealthProbeAddrress string `json:"healthProbeAddrress" yaml:"healthProbeAddrress"`
	MetricsAddress      string `json:"metricsAddress" yaml:"metricsAddress"`
	Address             string `json:"address" yaml:"address"`
	CertPath            string `json:"certPath" yaml:"certPath"`
	KeyPath             string `json:"keyPath" yaml:"keyPath"`

	// WebhookPort/WebhookCertPath/WebhookKeyPath configure a second,
	// independent TLS listener serving only the CRD conversion webhook
	// (POST /convert - see echoserver.RegisterNexusRoutes). Kubernetes CRD
	// conversion webhooks are unconditionally HTTPS, but the main API
	// server above is typically plain HTTP (CertPath/KeyPath empty), so
	// this listener is kept separate rather than forcing the whole REST
	// API onto TLS. All three are optional: when unset (the common case -
	// a deployment with no multi-version Kinds), no webhook listener is
	// started.
	WebhookPort     string `json:"webhookPort" yaml:"webhookPort,omitempty"`
	WebhookCertPath string `json:"webhookCertPath" yaml:"webhookCertPath,omitempty"`
	WebhookKeyPath  string `json:"webhookKeyPath" yaml:"webhookKeyPath,omitempty"`
}

var Cfg *Config

func LoadConfig(configFile string) (*Config, error) {
	config := &Config{}

	if configFile == "" {
		isPresent := false
		configFile, isPresent = os.LookupEnv("APIGWCONFIG")
		if !isPresent {
			configFile = APIGwConfigFileDefaullt
		}
	}

	file, err := os.Open(configFile)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %w", err)
	}
	configStr, err := io.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}
	err = yaml.Unmarshal(configStr, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Server.Address == "" {
		return config, fmt.Errorf("config doesn't contain Server.Address")
	}

	if config.Server.CertPath == "" {
		return config, fmt.Errorf("config doesn't contain Server.CertPath")
	}

	if config.Server.KeyPath == "" {
		return config, fmt.Errorf("config doesn't contain Server.KeyPath")
	}

	return config, nil
}
