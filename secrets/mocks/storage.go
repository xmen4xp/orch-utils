// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/open-edge-platform/orch-utils/secrets"
)

type StorageService struct {
	mock.Mock
}

var _ secrets.StorageService = &StorageService{}

// Put is a mock implementation.
func (m *StorageService) Put(ctx context.Context, namespace string, name string, values map[string]string) error {
	args := m.Called(ctx, namespace, name, values)
	return args.Error(0)
}

// Get is a mock implementation.
func (m *StorageService) Get(ctx context.Context, namespace string, name string) (map[string]string, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(map[string]string), args.Error(1)
}

// Delete is a mock implementation.
func (m *StorageService) Delete(ctx context.Context, namespace string, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}
