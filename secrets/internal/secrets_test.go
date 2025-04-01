// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/open-edge-platform/orch-utils/secrets"
	"github.com/open-edge-platform/orch-utils/secrets/internal"
	"github.com/open-edge-platform/orch-utils/secrets/mocks"
)

var _ = Describe("Secrets Provider Initialization", func() {
	var (
		ctx                context.Context
		log                *zap.SugaredLogger
		secretsProviderSvc *mocks.ProviderService
		storageSvc         *mocks.StorageService
		ts                 *httptest.Server
	)

	BeforeEach(func() {
		ctx = context.Background()

		zapConfig := zap.NewDevelopmentConfig()
		logger, err := zapConfig.Build()
		Expect(err).ToNot(HaveOccurred())
		log = logger.Sugar()

		secretsProviderSvc = &mocks.ProviderService{}
		storageSvc = &mocks.StorageService{}

		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprintln(w, "I am a mock OIDC IdP server")
		}))
	})

	AfterEach(func() {
		log.Sync() //nolint: errcheck
		ts.Close()
	})

	Context("Secrets provider service is initialized", func() {
		It("should initialize the secrets provider service and store the keys", func() {
			secretsProviderSvc.On("Initialized").Return(false, nil)
			secretsProviderSvc.On("Initialize", mock.AnythingOfType("context.backgroundCtx")).Return("", nil)
			storageSvc.On(
				"Put",
				mock.Anything,
				"orch-platform",
				internal.VaultKeysKubernetesSecretName,
				map[string]string{
					internal.VaultKeysKubernetesSecretName: "",
				},
			).Return(nil)
			storageSvc.On(
				"Get",
				mock.Anything,
				"orch-platform",
				internal.VaultKeysKubernetesSecretName,
			).Return(
				map[string]string{
					internal.VaultKeysKubernetesSecretName: `{ "root_token": "mock-token" }`,
				},
				nil,
			)
			secretsProviderSvc.On("SetToken", mock.Anything).Return()
			secretsProviderSvc.On("CreateOrchSvcSecretsStore").Return(nil)
			secretsProviderSvc.On("CreateOIDCAuth").Return(nil)

			err := internal.Configure(
				ctx,
				log,
				&secrets.Config{AuthOIDCIdPDiscoveryURL: ts.URL, AutoInit: true},
				secretsProviderSvc,
				storageSvc,
			)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
