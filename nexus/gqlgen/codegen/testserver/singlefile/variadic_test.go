// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package singlefile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/client"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql/handler"
)

func TestVariadic(t *testing.T) {
	resolver := &Stub{}
	resolver.QueryResolver.VariadicModel = func(ctx context.Context) (*VariadicModel, error) {
		return &VariadicModel{}, nil
	}
	c := client.New(handler.NewDefaultServer(
		NewExecutableSchema(Config{Resolvers: resolver}),
	))

	var resp struct {
		VariadicModel struct {
			Value string
		}
	}
	err := c.Post(`query { variadicModel { value(rank: 1) } }`, &resp)
	require.NoError(t, err)
	require.Equal(t, resp.VariadicModel.Value, "1")

	err = c.Post(`query { variadicModel { value(rank: 2) } }`, &resp)
	require.NoError(t, err)
	require.Equal(t, resp.VariadicModel.Value, "2")
}
