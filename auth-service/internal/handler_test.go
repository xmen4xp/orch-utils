// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-edge-platform/orch-utils/auth-service/internal"
)

var expectedStaticClaimRole = []string{`realm_access.roles.#(=="en-agent-rw")`}

var _ = Describe("Auth service with RBAC", func() {
	var roles *internal.RoleStore
	BeforeEach(func() {
		roles = internal.NewRoleStore(expectedStaticClaimRole)
	})
	Context("Auth service static roles", func() {
		It("should return 200 (OK) status code when using valid token with a valid static role", func() {
			tk, err := genToken(true, false, false)
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)
			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)
			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusOK))
		})

		It("should return 403 (Forbidden) status code when using valid token without a valid claim", func() {
			tk, err := genToken(false, false, false)
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)

			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusForbidden))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid claims"))
		})

		It("should return 401 (Unauthorized) when Auth header is missing in the request", func() {
			tk, err := genToken(false, false, false)
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)

			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid token"))
		})

		It("should return 401 (Unauthorized) when invalid token string is provided", func() {
			tk, err := genToken(false, false, false)
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)

			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer xyz123-token")

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid token"))
		})

		It("should return 401 (Unauthorized) when using expired token", func() {
			tk, err := genToken(true, true, false)
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)
			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid token"))
		})

		It("should return 401 (Unauthorized) when token is signed by unknown PKI", func() {
			tk, err := genToken(true, false, true)
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)
			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid token"))
		})
	})
})

var expectedDynamicClaimRole = []string{`realm_access.roles.#(=="{projectId}_en-agent-rw")`}

var _ = Describe("Auth service with RBAC", func() {
	var roles *internal.RoleStore
	BeforeEach(func() {
		roles = internal.NewRoleStore(expectedDynamicClaimRole)
		roles.SetProjectIDs([]string{"project1", "project2"})
		roles.UpdateDynamicRoles()
	})

	Context("Auth service dynamic roles", func() {
		It("should return 200 (OK) status code when using valid token with a valid role", func() {
			tk, err := genToken(true, false, false, []string{"project1"})
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)
			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)
			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusOK))
		})

		It("should return 403 (Forbidden) status code when using valid token without a valid claim", func() {
			tk, err := genToken(true, false, false, []string{"invalid"})
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, roles)
			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)
			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusForbidden))
		})

		It("should return 200 (OK) status code when using valid token with a valid role - added project", func() {
			tmpStore := internal.NewRoleStore(expectedDynamicClaimRole)

			// No project right now - should return 403
			tk, err := genToken(true, false, false, []string{"project1"})
			Expect(err).ToNot(HaveOccurred())

			handler := internal.NewHandler(tk.jwks, tmpStore)
			req, err := http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)
			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusForbidden))

			// Project gets added to the store
			tmpStore.SetProjectIDs([]string{"project1"})
			tmpStore.UpdateDynamicRoles()

			// Should return 200 now
			handler = internal.NewHandler(tk.jwks, roles)
			req, err = http.NewRequest(http.MethodGet, "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)
			rr = httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusOK))
		})
	})
})

type tokenKey struct {
	token string
	jwks  jwk.Set
}

func genToken(grantAccess bool, isExpired bool, isSignedByUnknownPKI bool, projectNames ...[]string) (tokenKey, error) {
	// Create a new token object.
	token := jwt.New(jwt.SigningMethodRS256)

	roles := []string{
		"admin",
		"test-role",
		"test-role2",
	}
	if grantAccess {
		// this is the expected role for accessing auth service.
		roles = append(roles, "en-agent-rw")
	}

	if len(projectNames) != 0 {
		for _, project := range projectNames {
			for _, name := range project {
				roles = append(roles, name+"_en-agent-rw")
			}
		}
	}

	realmAccess := map[string][]string{
		"roles": roles,
	}

	// Set some claims.
	if isExpired {
		token.Claims = jwt.MapClaims{
			"exp":          time.Now().Add(time.Hour * (-72)).Unix(),
			"realm_access": realmAccess,
		}
	} else {
		token.Claims = jwt.MapClaims{
			"exp":          time.Now().Add(time.Hour * 72).Unix(),
			"realm_access": realmAccess,
		}
	}

	// Generate key pair.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tokenKey{}, err
	}
	// Sign and get the complete encoded token as a string.
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return tokenKey{}, err
	}
	pubKey := &privateKey.PublicKey
	if isSignedByUnknownPKI {
		// generate another key pair and return public key which should fail verification
		privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return tokenKey{}, err
		}
		pubKey = &privateKey2.PublicKey
	}

	// Create a JWK from the RSA public key
	jwkKey, err := jwk.FromRaw(pubKey)
	if err != nil {
		return tokenKey{}, fmt.Errorf("failed to create JWK: %w", err)
	}

	// Create a JWKS with the JWK
	jwks := jwk.NewSet()
	if err := jwks.AddKey(jwkKey); err != nil {
		return tokenKey{}, err
	}
	return tokenKey{
		token: tokenString,
		jwks:  jwks,
	}, nil
}
