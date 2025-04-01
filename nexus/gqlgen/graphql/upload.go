// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"fmt"
	"io"
)

type Upload struct {
	File        io.ReadSeeker
	Filename    string
	Size        int64
	ContentType string
}

func MarshalUpload(f Upload) Marshaler {
	return WriterFunc(func(w io.Writer) {
		io.Copy(w, f.File)
	})
}

func UnmarshalUpload(v interface{}) (Upload, error) {
	upload, ok := v.(Upload)
	if !ok {
		return Upload{}, fmt.Errorf("%T is not an Upload", v)
	}
	return upload, nil
}
