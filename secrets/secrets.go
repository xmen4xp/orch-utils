// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package secrets

import "context"

// ProviderService stores secrets to a backing provider.
type ProviderService interface {
	// Initialized returns true if the secrets provider is already initialized.
	Initialized() (bool, error)
	// Initialize initializes the secrets provider and returns the encryption keys and CA certificate.
	Initialize(ctx context.Context) (string, error)
	SetToken(string)
	RevokeToken() error
	CreateOrchSvcSecretsStore() error
	CreateOIDCAuth() error
}

// StorageService stores data.
type StorageService interface {
	// Get retrieves values of name in the data store.
	Get(ctx context.Context, namespace string, name string) (map[string]string, error)
	// Put stores values at name in the data store.
	Put(ctx context.Context, namespace string, name string, values map[string]string) error
	// Delete removes by name in the data store.
	Delete(ctx context.Context, namespace string, name string) error
}
