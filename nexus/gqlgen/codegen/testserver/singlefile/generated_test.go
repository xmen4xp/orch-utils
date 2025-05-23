// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:generate rm -f resolver.go
//go:generate go run ../../../testdata/gqlgen.go -config gqlgen.yml -stub stub.go

package singlefile

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/client"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql/handler"
)

func TestForcedResolverFieldIsPointer(t *testing.T) {
	field, ok := reflect.TypeOf((*ForcedResolverResolver)(nil)).Elem().MethodByName("Field")
	require.True(t, ok)
	require.Equal(t, "*singlefile.Circle", field.Type.Out(0).String())
}

func TestEnums(t *testing.T) {
	t.Run("list of enums", func(t *testing.T) {
		require.Equal(t, StatusOk, AllStatus[0])
		require.Equal(t, StatusError, AllStatus[1])
	})

	t.Run("invalid enum values", func(t *testing.T) {
		require.Equal(t, StatusOk, AllStatus[0])
		require.Equal(t, StatusError, AllStatus[1])
	})
}

func TestUnionFragments(t *testing.T) {
	resolvers := &Stub{}
	resolvers.QueryResolver.ShapeUnion = func(ctx context.Context) (ShapeUnion, error) {
		return &Circle{Radius: 32}, nil
	}

	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolvers}))
	c := client.New(srv)

	t.Run("inline fragment on union", func(t *testing.T) {
		var resp struct {
			ShapeUnion struct {
				Radius float64
			}
		}
		c.MustPost(`query {
			shapeUnion {
				... on Circle {
					radius
				}
			}
		}
		`, &resp)
		require.NotEmpty(t, resp.ShapeUnion.Radius)
	})

	t.Run("named fragment", func(t *testing.T) {
		var resp struct {
			ShapeUnion struct {
				Radius float64
			}
		}
		c.MustPost(`query {
			shapeUnion {
				...C
			}
		}

		fragment C on ShapeUnion {
			... on Circle {
				radius
			}
		}
		`, &resp)
		require.NotEmpty(t, resp.ShapeUnion.Radius)
	})
}
