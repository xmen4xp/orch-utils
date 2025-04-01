// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/tidwall/gjson"
)

// NewFileHandler returns a handler func for accessing a file server which
// performs keycloak token verification including RBAC.
func NewFileHandler(keyset jwk.Set, fs http.Handler, allowRoles []string, emptyRSToken bool) http.HandlerFunc {
	if emptyRSToken {
		return func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte("anonymous"))
			if err == nil {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := jwt.ParseRequest(r,
			jwt.WithHeaderKey("Authorization"),
			jwt.WithKeySet(keyset, jws.WithRequireKid(false), jws.WithInferAlgorithmFromKey(true)),
			jwt.WithVerify(true),
			jwt.WithValidate(true),
		)
		if err != nil {
			log.Println("Invalid token while parsing it:", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		if err := verifyClaims(token, allowRoles); err != nil {
			log.Println("failed verifying claims:", err)
			http.Error(w, "Invalid claims", http.StatusForbidden)
			return
		}
		fs.ServeHTTP(w, r)
	}
}

func verifyClaims(token jwt.Token, allowRoles []string) error {
	payload, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("could not marshal claims: %v", err)
	}
	for _, role := range allowRoles {
		value := gjson.Get(string(payload), role)
		if len(value.String()) > 0 {
			return nil
		}
	}
	return fmt.Errorf("could not find expected roles: %v", allowRoles)
}
