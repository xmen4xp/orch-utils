// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package transport_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql/handler/testserver"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql/handler/transport"
)

func TestOptions(t *testing.T) {
	t.Run("responds to options requests with default methods", func(t *testing.T) {
		h := testserver.New()
		h.AddTransport(transport.Options{})
		resp := doRequest(h, "OPTIONS", "/graphql?query={me{name}}", ``)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "OPTIONS, GET, POST", resp.Header().Get("Allow"))
	})

	t.Run("responds to options requests with specified methods", func(t *testing.T) {
		h := testserver.New()
		h.AddTransport(transport.Options{
			AllowedMethods: []string{http.MethodOptions, http.MethodPost, http.MethodHead},
		})
		resp := doRequest(h, "OPTIONS", "/graphql?query={me{name}}", ``)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "OPTIONS, POST, HEAD", resp.Header().Get("Allow"))
	})

	t.Run("responds to head requests", func(t *testing.T) {
		h := testserver.New()
		h.AddTransport(transport.Options{})
		resp := doRequest(h, "HEAD", "/graphql?query={me{name}}", ``)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})
}
