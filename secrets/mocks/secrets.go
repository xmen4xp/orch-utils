// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/open-edge-platform/orch-utils/secrets"
)

type ProviderService struct {
	mock.Mock
}

var _ secrets.ProviderService = &ProviderService{}

// Initialized is a mock implementation.
func (m *ProviderService) Initialized() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

// Initialize is a mock implementation.
func (m *ProviderService) Initialize(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// SetToken is a mock implementation.
func (m *ProviderService) SetToken(t string) {
	m.Called(t)
}

// RevokeToken is a mock implementation.
func (m *ProviderService) RevokeToken() error {
	args := m.Called()
	return args.Error(0)
}

// CreateOrchSvcSecretsStore is a mock implementation.
func (m *ProviderService) CreateOrchSvcSecretsStore() error {
	args := m.Called()
	return args.Error(0)
}

// CreateOIDCAuth is a mock implementation.
func (m *ProviderService) CreateOIDCAuth() error {
	args := m.Called()
	return args.Error(0)
}
