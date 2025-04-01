// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

type RecoverFunc func(ctx context.Context, err interface{}) (userMessage error)

func DefaultRecover(ctx context.Context, err interface{}) error {
	fmt.Fprintln(os.Stderr, err)
	fmt.Fprintln(os.Stderr)
	debug.PrintStack()

	return gqlerror.Errorf("internal system error")
}
