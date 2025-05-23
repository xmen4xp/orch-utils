// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package policypkg

type RandomDescription struct {
	DiscriptionA string
	DiscriptionB string
	DiscriptionC string
	DiscriptionD string
}

type RandomStatus struct {
	StatusX int
	StatusY int
}

type RandomConst1 string
type RandomConst2 string
type RandomConst3 string

const (
	MyConst3 RandomConst3 = "Const3"
	MyConst2 RandomConst2 = "Const2"
	MyConst1 RandomConst1 = "Const1"
)
