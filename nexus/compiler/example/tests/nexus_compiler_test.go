// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package nexus_compiler_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNexusCompiler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nexus tests")
}
