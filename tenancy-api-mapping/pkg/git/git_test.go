/*
 * Copyright (C) 2025 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions
 * and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package git_test

import (
	"errors"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/open-edge-platform/orch-utils/tenancy-api-mapping/pkg/git"
)

// MockCmdRunner is a mock implementation of the CmdRunner interface.
type MockCmdRunner struct {
	ShouldFail bool
}

// RunCommand mocks the command execution.
func (m *MockCmdRunner) RunCommand(name, dir string, args ...string) error {
	fmt.Println(name, dir, args)
	if m.ShouldFail {
		return errors.New("command failed")
	}
	return nil
}

var _ = ginkgo.Describe("Git", func() {
	var (
		mockRunner *MockCmdRunner
		repoPath   string
	)

	ginkgo.BeforeEach(func() {
		mockRunner = &MockCmdRunner{}
		repoPath = "/tmp/repo"
	})

	ginkgo.Describe("InitSubmodule", func() {
		ginkgo.Context("when command execution succeeds", func() {
			ginkgo.It("should not return an error", func() {
				err := git.InitSubmodule(mockRunner, repoPath, "https://example.com/repo.git", "main", "submodule")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when command execution fails", func() {
			ginkgo.It("should return an error", func() {
				mockRunner.ShouldFail = true
				err := git.InitSubmodule(mockRunner, repoPath, "https://example.com/repo.git", "main", "submodule")
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})
})
