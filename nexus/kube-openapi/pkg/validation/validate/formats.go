// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"reflect"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/strfmt"
)

type formatValidator struct {
	Format       string
	Path         string
	In           string
	KnownFormats strfmt.Registry
}

func (f *formatValidator) SetPath(path string) {
	f.Path = path
}

func (f *formatValidator) Applies(source interface{}, kind reflect.Kind) bool {
	doit := func() bool {
		if source == nil {
			return false
		}
		switch source := source.(type) {
		case *spec.Schema:
			return kind == reflect.String && f.KnownFormats.ContainsName(source.Format)
		}
		return false
	}
	r := doit()
	debugLog("format validator for %q applies %t for %T (kind: %v)\n", f.Path, r, source, kind)
	return r
}

func (f *formatValidator) Validate(val interface{}) *Result {
	result := new(Result)
	debugLog("validating \"%v\" against format: %s", val, f.Format)

	if err := FormatOf(f.Path, f.In, f.Format, val.(string), f.KnownFormats); err != nil {
		result.AddErrors(err)
	}

	if result.HasErrors() {
		return result
	}
	return nil
}
