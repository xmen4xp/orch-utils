// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"nexus-api-gw/pkg/config"

	"github.com/labstack/echo/v4"
)

// graphQLPathPrefix is the external path under which the datamodel GraphQL API
// is exposed through the API gateway. The generated GraphQL server serves the
// playground at "/" and the query endpoint at "/query"; this prefix is stripped
// before the request is forwarded to the backend service.
const graphQLPathPrefix = "/apis/graphql/v1"

// setupGraphQLProxy registers a reverse proxy from /apis/graphql/v1/* to the
// nexus-graphql backend service. Because the route is more specific than the
// "/apis/*" kube proxy catch-all, Echo's router matches it preferentially, so
// GraphQL requests are routed to the GraphQL server while all other /apis/*
// traffic continues to the Kubernetes API server.
//
// It is a no-op when no GraphQL backend is configured.
func setupGraphQLProxy(e *echo.Echo, conf *config.Config) {
	if conf == nil || conf.GraphQLBackend == "" {
		return
	}

	target, err := url.Parse(conf.GraphQLBackend)
	if err != nil {
		log.Warn().Msgf("Could not parse GraphQL backend URL %q: %v", conf.GraphQLBackend, err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	baseDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		baseDirector(req)
		// Strip the external prefix so the backend receives "/query" or "/".
		trimmed := strings.TrimPrefix(req.URL.Path, graphQLPathPrefix)
		if trimmed == "" {
			trimmed = "/"
		}
		req.URL.Path = trimmed
		req.Host = target.Host
	}

	handler := echo.WrapHandler(proxy)
	e.Any(graphQLPathPrefix, handler)
	e.Any(graphQLPathPrefix+"/*", handler)
	log.Info().Msgf("Registered GraphQL proxy %s/* -> %s", graphQLPathPrefix, conf.GraphQLBackend)
}
