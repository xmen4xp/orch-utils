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

func TestModelMethods(t *testing.T) {
	resolver := &Stub{}
	resolver.QueryResolver.ModelMethods = func(ctx context.Context) (methods *ModelMethods, e error) {
		return &ModelMethods{}, nil
	}
	resolver.ModelMethodsResolver.ResolverField = func(ctx context.Context, obj *ModelMethods) (b bool, e error) {
		return true, nil
	}

	c := client.New(handler.NewDefaultServer(
		NewExecutableSchema(Config{Resolvers: resolver}),
	))
	t.Run("without context", func(t *testing.T) {
		var resp struct {
			ModelMethods struct {
				NoContext bool
			}
		}
		err := c.Post(`query { modelMethods{ noContext } }`, &resp)
		require.NoError(t, err)
		require.True(t, resp.ModelMethods.NoContext)
	})
	t.Run("with context", func(t *testing.T) {
		var resp struct {
			ModelMethods struct {
				WithContext bool
			}
		}
		err := c.Post(`query { modelMethods{ withContext } }`, &resp)
		require.NoError(t, err)
		require.True(t, resp.ModelMethods.WithContext)
	})
}
