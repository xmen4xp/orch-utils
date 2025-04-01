// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"reflect"
	"testing"
)

func TestCanonicalName(t *testing.T) {

	var tests = []struct {
		input    string
		expected string
	}{
		{"k8s.io/api/core/v1.Pod", "io.k8s.api.core.v1.Pod"},
		{"k8s.io/api/networking/v1/NetworkPolicy", "io.k8s.api.networking.v1.NetworkPolicy"},
		{"k8s.io/api/apps/v1beta2.Scale", "io.k8s.api.apps.v1beta2.Scale"},
		{"servicecatalog.k8s.io/foo/bar/v1alpha1.Baz", "io.k8s.servicecatalog.foo.bar.v1alpha1.Baz"},
	}
	for _, test := range tests {
		if got := ToRESTFriendlyName(test.input); got != test.expected {
			t.Errorf("ToRESTFriendlyName(%q) = %v", test.input, got)
		}
	}
}

type TestType struct{}

func TestGetCanonicalTypeName(t *testing.T) {

	var tests = []struct {
		input    interface{}
		expected string
	}{
		{TestType{}, "github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util.TestType"},
		{&TestType{}, "github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util.TestType"},
	}
	for _, test := range tests {
		if got := GetCanonicalTypeName(test.input); got != test.expected {
			t.Errorf("GetCanonicalTypeName(%q) = %v", reflect.TypeOf(test.input), got)
		}
	}
}
