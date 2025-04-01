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
	"testing/fstest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/open-edge-platform/orch-utils/token-fs/internal"
)

const expectedClaimRole = `realm_access.roles.#(=="release-service-access-token-read-role")`

var _ = Describe("File Server with jwt RBAC", func() {
	var (
		dataFS fstest.MapFS
		roles  []string
	)
	BeforeEach(func() {
		// Create a map-based file system
		dataFS = fstest.MapFS{
			"token": &fstest.MapFile{
				Data: []byte("rs-token-string"),
			},
		}
		roles = []string{expectedClaimRole}
	})
	Context("Token FS", func() {
		It("should return the Release Service token string when using a valid input token", func() {
			tk, err := genToken(true, false, false)
			Expect(err).ToNot(HaveOccurred())

			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, false)
			req, err := http.NewRequest("GET", "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(ContainSubstring("rs-token-string"))
		})

		It("should return 403 (Forbidden) status code when using valid token without a valid claim", func() {
			tk, err := genToken(false, false, false)
			Expect(err).ToNot(HaveOccurred())

			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, false)

			req, err := http.NewRequest("GET", "/token", nil)
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

			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, false)

			req, err := http.NewRequest("GET", "/token", nil)
			Expect(err).ToNot(HaveOccurred())

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid token"))
		})

		It("should return 401 (Unauthorized) when invalid token string is provided", func() {
			tk, err := genToken(false, false, false)
			Expect(err).ToNot(HaveOccurred())

			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, false)

			req, err := http.NewRequest("GET", "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer xyz123-token")

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid token"))
		})

		It("should return 404 (not found) status code when requesting non-existing file", func() {
			tk, err := genToken(true, false, false)
			Expect(err).ToNot(HaveOccurred())

			// This test is not required from the code coverage statndpoint but
			// it demonstrates the expected response code from the file server
			// when requesting a non-existing file path.
			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, false)

			req, err := http.NewRequest("GET", "/bad-file", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusNotFound))
			Expect(rr.Body.String()).To(ContainSubstring("page not found"))
		})

		It("should return 401 (Unauthorized) when using expired token", func() {
			tk, err := genToken(true, true, false)
			Expect(err).ToNot(HaveOccurred())

			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, false)
			req, err := http.NewRequest("GET", "/token", nil)
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

			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, false)
			req, err := http.NewRequest("GET", "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(rr.Body.String()).To(ContainSubstring("Invalid token"))
		})

		It("should return the empty token when emptyRSToken is true", func() {
			tk, err := genToken(true, false, false)
			Expect(err).ToNot(HaveOccurred())

			fs := http.FileServer(http.FS(dataFS))
			handler := internal.NewFileHandler(tk.jwks, fs, roles, true)
			req, err := http.NewRequest("GET", "/token", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+tk.token)

			rr := httptest.NewRecorder()
			handler(rr, req)
			Expect(rr.Result().StatusCode).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal("anonymous"))
		})
	})
})

type tokenKey struct {
	token string
	jwks  jwk.Set
}

func genToken(grantAccess bool, isExpired bool, isSignedByUnknownPKI bool) (tokenKey, error) {
	// Create a new token object.
	token := jwt.New(jwt.SigningMethodRS256)

	roles := []string{
		"infra-manager-core-write-role",
		"catalog-restricted-read-role",
		"admin",
		"app-service-proxy-read-role",
		"clusters-write-role",
		"offline_access",
		"app-resource-manager-write-role",
		"cluster-templates-write-role",
		"clusters-read-role",
		"uma_authorization",
		"alert-definitions-read-role",
		"secrets-pki-readwrite-role",
		"cluster-artifacts-read-role",
		"catalog-restricted-write-role",
		"app-deployment-manager-read-role",
		"lp-read-only-role",
		"default-roles-master",
		"alerts-read-role",
		"catalog-publisher-read-role",
		"alert-definitions-write-role",
		"app-vm-console-write-role",
		"lp-admin-role",
		"app-deployment-manager-write-role",
		"cluster-templates-read-role",
		"create-realm",
		"catalog-other-read-role",
		"cluster-artifacts-write-role",
		"catalog-other-write-role",
		"infra-manager-core-read-role",
		"catalog-publisher-write-role",
		"alert-receivers-write-role",
		"lp-read-write-role",
		"alert-receivers-read-role",
		"app-service-proxy-write-role",
		"app-resource-manager-read-role",
	}
	if grantAccess {
		// this is the expected role for accessing RS token.
		roles = append(roles, "release-service-access-token-read-role")
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
		return tokenKey{}, fmt.Errorf("Failed to create JWK: %v", err)
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
