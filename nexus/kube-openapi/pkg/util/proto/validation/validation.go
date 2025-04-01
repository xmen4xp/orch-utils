// Copyright 2017 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util/proto"
)

func ValidateModel(obj interface{}, schema proto.Schema, name string) []error {
	rootValidation, err := itemFactory(proto.NewPath(name), obj)
	if err != nil {
		return []error{err}
	}
	schema.Accept(rootValidation)
	return rootValidation.Errors()
}
