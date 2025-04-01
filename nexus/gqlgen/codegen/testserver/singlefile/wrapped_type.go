// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package singlefile

import "github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/codegen/testserver/singlefile/otherpkg"

type (
	WrappedScalar = otherpkg.Scalar
	WrappedStruct otherpkg.Struct
	WrappedMap    otherpkg.Map
	WrappedSlice  otherpkg.Slice
)
