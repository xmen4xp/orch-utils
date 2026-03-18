// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver

import (
	"net/http"
	"testing"

	"nexus-api-gw/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestGetHeaderValue_WithAlias(t *testing.T) {
	// Setup config with header aliases
	config.Cfg = &config.Config{
		HeaderAliases: map[string]string{
			"orgs.Org":         "x-org-id",
			"projects.Project": "x-project-id",
		},
	}
	defer func() { config.Cfg = nil }()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("x-org-id", "my-org")

	// Should return value from alias header
	result := GetHeaderValue(req, "orgs.Org")
	assert.Equal(t, "my-org", result)
}

func TestGetHeaderValue_FallbackToPrimaryName(t *testing.T) {
	// Setup config with header aliases
	config.Cfg = &config.Config{
		HeaderAliases: map[string]string{
			"orgs.Org": "x-org-id",
		},
	}
	defer func() { config.Cfg = nil }()

	req, _ := http.NewRequest("GET", "/test", nil)
	// Alias header not set, but primary name is set
	req.Header.Set("orgs.Org", "my-org-primary")

	// Should fall back to primary header name
	result := GetHeaderValue(req, "orgs.Org")
	assert.Equal(t, "my-org-primary", result)
}

func TestGetHeaderValue_AliasPreferredOverPrimary(t *testing.T) {
	// Setup config with header aliases
	config.Cfg = &config.Config{
		HeaderAliases: map[string]string{
			"orgs.Org": "x-org-id",
		},
	}
	defer func() { config.Cfg = nil }()

	req, _ := http.NewRequest("GET", "/test", nil)
	// Both alias and primary headers are set
	req.Header.Set("x-org-id", "alias-org")
	req.Header.Set("orgs.Org", "primary-org")

	// Alias should be preferred
	result := GetHeaderValue(req, "orgs.Org")
	assert.Equal(t, "alias-org", result)
}

func TestGetHeaderValue_NoAliasConfigured(t *testing.T) {
	// Setup config without header aliases
	config.Cfg = &config.Config{
		HeaderAliases: nil,
	}
	defer func() { config.Cfg = nil }()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("orgs.Org", "my-org")

	// Should use primary header name
	result := GetHeaderValue(req, "orgs.Org")
	assert.Equal(t, "my-org", result)
}

func TestGetHeaderValue_NilConfig(t *testing.T) {
	// No config set
	config.Cfg = nil

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("orgs.Org", "my-org")

	// Should use primary header name without panic
	result := GetHeaderValue(req, "orgs.Org")
	assert.Equal(t, "my-org", result)
}

func TestGetHeaderValue_EmptyAliasHeader(t *testing.T) {
	// Setup config with header aliases
	config.Cfg = &config.Config{
		HeaderAliases: map[string]string{
			"orgs.Org": "x-org-id",
		},
	}
	defer func() { config.Cfg = nil }()

	req, _ := http.NewRequest("GET", "/test", nil)
	// Alias header is set but empty, primary has value
	req.Header.Set("x-org-id", "")
	req.Header.Set("orgs.Org", "primary-org")

	// Should fall back to primary when alias is empty
	result := GetHeaderValue(req, "orgs.Org")
	assert.Equal(t, "primary-org", result)
}
