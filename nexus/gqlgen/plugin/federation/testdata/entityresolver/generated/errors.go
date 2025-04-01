// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package generated

import "errors"

// Errors defined for retained code that we want to stick around between generations.
var (
	ErrResolvingHelloWithErrorsByName         = errors.New("error resolving HelloWithErrorsByName")
	ErrEmptyKeyResolvingHelloWithErrorsByName = errors.New("error (empty key) resolving HelloWithErrorsByName")
)
