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

func TestDefaultScalarImplementation(t *testing.T) {
	resolvers := &Stub{}

	c := client.New(handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolvers})))

	resolvers.QueryResolver.DefaultScalar = func(ctx context.Context, arg string) (i string, e error) {
		return arg, nil
	}

	t.Run("with arg value", func(t *testing.T) {
		var resp struct{ DefaultScalar string }
		c.MustPost(`query { defaultScalar(arg: "fff") }`, &resp)
		require.Equal(t, "fff", resp.DefaultScalar)
	})

	t.Run("with default value", func(t *testing.T) {
		var resp struct{ DefaultScalar string }
		c.MustPost(`query { defaultScalar  }`, &resp)
		require.Equal(t, "default", resp.DefaultScalar)
	})
}
