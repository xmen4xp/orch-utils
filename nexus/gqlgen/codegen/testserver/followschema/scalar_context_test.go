// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package followschema

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/client"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql/handler"
)

func TestFloatInfAndNaN(t *testing.T) {
	resolvers := &Stub{}

	c := client.New(handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolvers})))

	resolvers.QueryResolver.Infinity = func(ctx context.Context) (float64, error) {
		return math.Inf(-1), nil
	}

	t.Run("errors on marshaller with context", func(t *testing.T) {
		err := c.Post(`query { infinity }`, nil)
		require.Error(t, err)
	})
}

func TestContextPassedToMarshal(t *testing.T) {
	resolvers := &Stub{}

	c := client.New(handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolvers})))

	resolvers.QueryResolver.StringFromContextInterface = func(ctx context.Context) (*StringFromContextInterface, error) {
		return &StringFromContextInterface{}, nil
	}
	resolvers.QueryResolver.StringFromContextFunction = func(ctx context.Context) (string, error) {
		return "", nil
	}

	var res struct {
		StringFromContextInterface string
		StringFromContextFunction  string
	}
	err := c.Post(`query my_name {
		stringFromContextInterface
		stringFromContextFunction
	}`, &res)
	require.NoError(t, err)
	require.Equal(t, "stringFromContextInterface", res.StringFromContextInterface)
	require.Equal(t, "stringFromContextFunction", res.StringFromContextFunction)
}
