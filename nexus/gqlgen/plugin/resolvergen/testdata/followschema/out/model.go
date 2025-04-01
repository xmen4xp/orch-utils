// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package customresolver

import "context"

type Resolver struct{}

type QueryResolver interface {
	Resolver(ctx context.Context) (*Resolver, error)
}

type ResolverResolver interface {
	Name(ctx context.Context, obj *Resolver) (string, error)
}
