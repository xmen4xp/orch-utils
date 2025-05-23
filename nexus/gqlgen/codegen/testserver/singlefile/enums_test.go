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

func TestEnumsResolver(t *testing.T) {
	resolvers := &Stub{}
	resolvers.QueryResolver.EnumInInput = func(ctx context.Context, input *InputWithEnumValue) (EnumTest, error) {
		return input.Enum, nil
	}

	c := client.New(handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolvers})))

	t.Run("input with valid enum value", func(t *testing.T) {
		var resp struct {
			EnumInInput EnumTest
		}
		c.MustPost(`query {
			enumInInput(input: {enum: OK})
		}
		`, &resp)
		require.Equal(t, resp.EnumInInput, EnumTestOk)
	})

	t.Run("input with invalid enum value", func(t *testing.T) {
		var resp struct {
			EnumInInput EnumTest
		}
		err := c.Post(`query {
			enumInInput(input: {enum: INVALID})
		}
		`, &resp)
		require.EqualError(t, err, `http 422: {"errors":[{"message":"Value \"INVALID\" does not exist in \"EnumTest!\" enum.","locations":[{"line":2,"column":30}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}`)
	})

	t.Run("input with invalid enum value via vars", func(t *testing.T) {
		var resp struct {
			EnumInInput EnumTest
		}
		err := c.Post(`query ($input: InputWithEnumValue) {
			enumInInput(input: $input)
		}
		`, &resp, client.Var("input", map[string]interface{}{"enum": "INVALID"}))
		require.EqualError(t, err, `http 422: {"errors":[{"message":"INVALID is not a valid EnumTest","path":["variable","input","enum"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}`)
	})
}
