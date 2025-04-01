// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package vaults

import (
	"context"
	"fmt"
	"os"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/kubernetes"

	"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/pkg/log"
)

//nolint:revive // keep SecretsService name
type SecretsService interface {
	// ReadSecret reads a persistent secret under the given path and returns a stored object.
	// A consumer is responsible for parsing the returned object and converting it to an expected format.
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	// Logout terminates a user session. Should be always invoked after all operations are done.
	Logout(ctx context.Context)
}

var SecretServiceFactory = newVaultService

const (
	DefaultTimeout = 3 * time.Second

	vaultSecretBaseURL = `/secret/data/`

	DefaultVaultRole = "orch-svc"
	DefaultVaultURL  = "http://vault.orch-platform.svc.cluster.local:8200"

	EnvNameVaultURL     = "VAULT_URL"
	EnvNameVaultPKIRole = "VAULT_PKI_ROLE"
)

// VaultAPI wraps vault under interface to enable mocking for unit testing.
type VaultAPI interface {
	Read(ctx context.Context, path string) (*vault.Secret, error)
	RevokeToken(ctx context.Context) error
}

type vaultAPI struct {
	vaultClient *vault.Client
}

func (v vaultAPI) Read(ctx context.Context, path string) (*vault.Secret, error) {
	return v.vaultClient.Logical().ReadWithContext(ctx, path)
}

func (v vaultAPI) RevokeToken(ctx context.Context) error {
	// token can be left empty, see lib docs
	return v.vaultClient.Auth().Token().RevokeSelfWithContext(ctx, "")
}

type vaultService struct {
	vaultClient VaultAPI
}

func newVaultService(ctx context.Context) (SecretsService, error) {
	vaultURL := os.Getenv(EnvNameVaultURL)
	if vaultURL == "" {
		log.Warnf("%s env variable is not set, using default value", EnvNameVaultURL)
		vaultURL = DefaultVaultURL
	}

	vaultRole := os.Getenv(EnvNameVaultPKIRole)
	if vaultRole == "" {
		log.Warnf("%s env variable is not set, using default value", EnvNameVaultPKIRole)
		vaultRole = DefaultVaultRole
	}

	ss := &vaultService{}
	err := ss.login(ctx, vaultURL, vaultRole)
	if err != nil {
		return nil, err
	}
	return ss, err
}

func getVaultClient(vaultURL string) (*vault.Client, error) {
	config := vault.DefaultConfig()
	config.Address = vaultURL

	client, err := vault.NewClient(config)
	if err != nil {
		returnErr := fmt.Errorf("failed to create Vault client")
		log.Errorf("%s", err)
		return nil, returnErr
	}

	return client, nil
}

func loginToVault(ctx context.Context, vaultCli *vault.Client, vaultRole string) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	k8sAuth, err := kubernetes.NewKubernetesAuth(vaultRole)
	if err != nil {
		returnErr := fmt.Errorf("failed to create K8s auth credentials")
		log.Errorf("%s", err)
		return returnErr
	}

	authInfo, err := vaultCli.Auth().Login(ctx, k8sAuth)
	if err != nil {
		returnErr := fmt.Errorf("failed to login to Vault")
		log.Errorf("%s", err)
		return returnErr
	}

	if authInfo == nil {
		returnErr := fmt.Errorf("no auth info was returned after login to Vault")
		log.Errorf("%s", err)
		return returnErr
	}

	return nil
}

func (v *vaultService) login(ctx context.Context, vaultURL, vaultRole string) error {
	client, err := getVaultClient(vaultURL)
	if err != nil {
		return err
	}
	err = loginToVault(ctx, client, vaultRole)
	if err != nil {
		return err
	}
	v.vaultClient = &vaultAPI{client}
	return nil
}

func (v *vaultService) ReadSecret(ctx context.Context, secretName string) (map[string]interface{}, error) {
	secret, err := v.vaultClient.Read(ctx, vaultSecretBaseURL+secretName)
	if err != nil {
		returnErr := fmt.Errorf("failed to read secret from Vault")
		log.Errorf("%s", err)
		return nil, returnErr
	}

	// There are scenarios in which secret will be nil, even if there is no error.
	// For example, this can happen in the case of 204 No Content response.
	// See: https://github.com/hashicorp/vault/issues/18836
	if secret == nil {
		return nil, fmt.Errorf("secret %s not found", secretName)
	}

	return secret.Data, nil
}

func (v *vaultService) Logout(ctx context.Context) {
	err := v.vaultClient.RevokeToken(ctx)
	if err != nil {
		log.Errorf("Failed to log out from Vault: %v", err)
	}
}
