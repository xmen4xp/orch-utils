// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql"
)

// EnableIntrospection enables clients to reflect all of the types available on the graph.
type Introspection struct{}

var _ interface {
	graphql.OperationContextMutator
	graphql.HandlerExtension
} = Introspection{}

func (c Introspection) ExtensionName() string {
	return "Introspection"
}

func (c Introspection) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (c Introspection) MutateOperationContext(ctx context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	rc.DisableIntrospection = false
	return nil
}
