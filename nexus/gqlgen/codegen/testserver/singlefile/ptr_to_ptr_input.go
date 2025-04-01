// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package singlefile

type PtrToPtrOuter struct {
	Name        string
	Inner       *PtrToPtrInner
	StupidInner *******PtrToPtrInner
}

type PtrToPtrInner struct {
	Key   string
	Value string
}

type UpdatePtrToPtrOuter struct {
	Name        *string
	Inner       **UpdatePtrToPtrInner
	StupidInner ********UpdatePtrToPtrInner
}

type UpdatePtrToPtrInner struct {
	Key   *string
	Value *string
}
