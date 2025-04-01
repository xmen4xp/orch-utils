// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"encoding/json"
	"io"
)

func MarshalAny(v interface{}) Marshaler {
	return WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			panic(err)
		}
	})
}

func UnmarshalAny(v interface{}) (interface{}, error) {
	return v, nil
}
