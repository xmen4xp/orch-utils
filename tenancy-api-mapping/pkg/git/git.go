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

package git

import (
	"os"
	"os/exec"
	"path/filepath"
)

// CmdRunner is an interface that allows for mocking out the command execution.
type CmdRunner interface {
	RunCommand(name string, dir string, args ...string) error
}

// ExecCmdRunner is an implementation of CmdRunner that actually runs commands.
type ExecCmdRunner struct{}

// RunCommand runs the given command with the provided arguments.
func (r *ExecCmdRunner) RunCommand(name, dir string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}

// InitSubmodule initializes a Git submodule using the provided CmdRunner.
func InitSubmodule(runner CmdRunner, repoPath, submoduleURL, tag, submodulePath string) error {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
			return err
		}
	}

	// Add the submodule
	if err := runner.RunCommand(
		"git",
		repoPath,
		"submodule",
		"add",
		"-f",
		submoduleURL,
		submodulePath,
	); err != nil {
		return err
	}

	// Navigate to the submodule directory
	submoduleFullPath := filepath.Join(repoPath, submodulePath)
	if err := runner.RunCommand("git", submoduleFullPath, "checkout", "tags/"+tag); err != nil {
		return err
	}

	// Update the submodule
	return runner.RunCommand("git", repoPath, "submodule", "update", "--init", "--recursive")
}
