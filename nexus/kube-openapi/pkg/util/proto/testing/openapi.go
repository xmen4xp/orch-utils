// Copyright 2017 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"io/ioutil"
	"os"
	"sync"

	openapi_v2 "github.com/google/gnostic/openapiv2"
)

// Fake opens and returns a openapi swagger from a file Path. It will
// parse only once and then return the same copy everytime.
type Fake struct {
	Path string

	once     sync.Once
	document *openapi_v2.Document
	err      error
}

// OpenAPISchema returns the openapi document and a potential error.
func (f *Fake) OpenAPISchema() (*openapi_v2.Document, error) {
	f.once.Do(func() {
		_, err := os.Stat(f.Path)
		if err != nil {
			f.err = err
			return
		}
		spec, err := ioutil.ReadFile(f.Path)
		if err != nil {
			f.err = err
			return
		}
		f.document, f.err = openapi_v2.ParseDocument(spec)
	})
	return f.document, f.err
}

type Empty struct{}

func (Empty) OpenAPISchema() (*openapi_v2.Document, error) {
	return nil, nil
}
