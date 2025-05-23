// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package followschema

import (
	"context"
	"io"
	"strconv"

	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql"
)

type StringFromContextInterface struct {
	OperationName string
}

var (
	_ graphql.ContextMarshaler   = StringFromContextInterface{}
	_ graphql.ContextUnmarshaler = (*StringFromContextInterface)(nil)
)

func (StringFromContextInterface) MarshalGQLContext(ctx context.Context, w io.Writer) error {
	io.WriteString(w, strconv.Quote(graphql.GetFieldContext(ctx).Field.Name))
	return nil
}

func (i *StringFromContextInterface) UnmarshalGQLContext(ctx context.Context, v interface{}) error {
	i.OperationName = graphql.GetFieldContext(ctx).Field.Name
	return nil
}

func MarshalStringFromContextFunction(v string) graphql.ContextMarshaler {
	return graphql.ContextWriterFunc(func(ctx context.Context, w io.Writer) error {
		io.WriteString(w, strconv.Quote(graphql.GetFieldContext(ctx).Field.Name))
		return nil
	})
}

func UnmarshalStringFromContextFunction(ctx context.Context, v interface{}) (string, error) {
	return graphql.GetFieldContext(ctx).Field.Name, nil
}
