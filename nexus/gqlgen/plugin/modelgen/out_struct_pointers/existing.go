// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package out_struct_pointers

type ExistingType struct {
	Name     *string              `json:"name"`
	Enum     *ExistingEnum        `json:"enum"`
	Int      ExistingInterface    `json:"int"`
	Existing *MissingTypeNullable `json:"existing"`
}

type ExistingModel struct {
	Name string
	Enum ExistingEnum
	Int  ExistingInterface
}

type ExistingInput struct {
	Name string
	Enum ExistingEnum
	Int  ExistingInterface
}

type ExistingEnum string

type ExistingInterface interface {
	IsExistingInterface()
}

type ExistingUnion interface {
	IsExistingUnion()
}
