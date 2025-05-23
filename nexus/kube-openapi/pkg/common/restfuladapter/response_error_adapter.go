// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package restfuladapter

import (
	"github.com/emicklei/go-restful"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/common"
)

var _ common.StatusCodeResponse = &ResponseErrorAdapter{}

// ResponseErrorAdapter adapts a restful.ResponseError to common.StatusCodeResponse.
type ResponseErrorAdapter struct {
	Err *restful.ResponseError
}

func (r *ResponseErrorAdapter) Message() string {
	return r.Err.Message
}

func (r *ResponseErrorAdapter) Model() interface{} {
	return r.Err.Model
}

func (r *ResponseErrorAdapter) Code() int {
	return r.Err.Code
}
