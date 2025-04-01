// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"context"
	"fmt"

	"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/pkg/auth/vaults"
	"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/pkg/log"
)

const (
	//nolint:gosec // hardcoded secrets need to handle in future.
	ktcCredentialsSecretName = "ktc-m2m-client-secret"
)

var (
	secretClientID     string
	secretClientSecret string
)

var inst = &secretService{}

// SecretService implements the interaction with the secrets storage (e.g., Vault).
type SecretService interface {
	// Init initializes the SecretService.
	// It should always be invoked at the very beginning, before other methods are used.
	Init(ctx context.Context) error
	// GetClientID obtains the `client_id` secret (in the UUID format) from the SecretService.
	GetClientID()
	// GetClientSecret obtains the `client_secret` secret from the SecretService.
	GetClientSecret()
}

type secretService struct{}

func Init(ctx context.Context) error {
	return inst.init(ctx)
}

func GetClientID() string {
	return secretClientID
}

func GetClientSecret() string {
	return secretClientSecret
}

func (ss *secretService) init(ctx context.Context) error {
	vaultS, err := vaults.SecretServiceFactory(ctx)
	if err != nil {
		return err
	}
	defer vaultS.Logout(ctx)

	credentials, err := vaultS.ReadSecret(ctx, ktcCredentialsSecretName)
	if err != nil {
		return err
	}

	dataMap, ok := credentials["data"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("cannot read credentials data from Vault secret")
		log.Errorf("%s", err)
		return err
	}

	_clientID, exists := dataMap["client_id"]
	if !exists {
		err = fmt.Errorf("failed to get client_id from secrets service")
		log.Errorf("%s", err)
		return err
	}
	clientID, ok := _clientID.(string)
	if !ok {
		err = fmt.Errorf("wrong format of client_id read from Vault, expected string, got %T", _clientID)
		log.Errorf("%s", err)
		return err
	}
	secretClientID = clientID

	_clientSecret, exists := dataMap["client_secret"]
	if !exists {
		err = fmt.Errorf("failed to get client_id from secrets service")
		log.Errorf("%s", err)
		return err
	}
	clientSecret, ok := _clientSecret.(string)
	if !ok {
		err = fmt.Errorf("wrong format of client_secret read from Vault, expected string, got %T", _clientSecret)
		log.Errorf("%s", err)
		return err
	}
	secretClientSecret = clientSecret

	log.Debugf("Secrets successfully initialized")

	return nil
}
