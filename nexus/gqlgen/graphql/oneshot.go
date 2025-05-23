// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package graphql

import "context"

func OneShot(resp *Response) ResponseHandler {
	var oneshot bool

	return func(context context.Context) *Response {
		if oneshot {
			return nil
		}
		oneshot = true

		return resp
	}
}
