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

package config_test

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/open-edge-platform/orch-utils/tenancy-api-mapping/pkg/config"
)

var _ = ginkgo.Describe("Config", func() {
	var tempDir string
	var err error
	const defaultFileMode fs.FileMode = 0o600

	ginkgo.BeforeEach(func() {
		tempDir, err = os.MkdirTemp("", "configTest")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	ginkgo.Describe("LoadConfig", func() {
		ginkgo.It("should load the config from a file", func() {
			configFile := filepath.Join(tempDir, "config.yaml")
			configData := []byte("apiVersion: v1\nkind: Config\nmetadata:\n  name: test-config")
			err := os.WriteFile(configFile, configData, defaultFileMode)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			cfg, err := config.LoadConfig(configFile)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(cfg).NotTo(gomega.BeNil())
			// Add more assertions to validate the contents of cfg
		})

		ginkgo.It("should return an error for a non-existent file", func() {
			_, err := config.LoadConfig("non-existent.yaml")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("ReadCRFile", func() {
		ginkgo.It("should read the content of a CR file", func() {
			crFile := filepath.Join(tempDir, "cr.yaml")
			crData := []byte("apiVersion: v1\nkind: APIMappingConfig\n...")
			err := os.WriteFile(crFile, crData, defaultFileMode)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			content, err := config.ReadCRFile(crFile)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(content).To(gomega.Equal(crData))
		})

		ginkgo.It("should return an error if the CR file does not exist", func() {
			_, err := config.ReadCRFile("non-existent-cr.yaml")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("ParseCRContent", func() {
		ginkgo.It("should parse the content of a CR file", func() {
			crData := []byte(`
apiVersion: apimappingconfig.edge-orchestrator.intel.com/v1
kind: APIMappingConfig
metadata:
  name: test-cr
spec:
  repoConf:
    url: "https://example.com/repo.git"
    tag: "main"
    specFilePath: "path/to/spec.yaml"
`)
			crConfig, err := config.ParseCRContent(crData)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(crConfig.APIVersion).To(gomega.Equal("apimappingconfig.edge-orchestrator.intel.com/v1"))
			gomega.Expect(crConfig.Kind).To(gomega.Equal("APIMappingConfig"))
			gomega.Expect(crConfig.Metadata.Name).To(gomega.Equal("test-cr"))
			gomega.Expect(crConfig.Spec.RepoConf.URL).To(gomega.Equal("https://example.com/repo.git"))
			// Add more assertions to validate the contents of crConfig
		})

		ginkgo.It("should return an error if the CR content is invalid", func() {
			crData := []byte("invalid-content")
			_, err := config.ParseCRContent(crData)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("GetCrYAMLFilePaths", func() {
		ginkgo.It("should return file paths of all CR YAML files in a directory", func() {
			crFile1 := filepath.Join(tempDir, "cr1.yaml")
			crFile2 := filepath.Join(tempDir, "cr2.yml")
			nonCrFile := filepath.Join(tempDir, "not-a-cr.txt")
			err := os.WriteFile(crFile1, []byte("content"), defaultFileMode)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.WriteFile(crFile2, []byte("content"), defaultFileMode)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.WriteFile(nonCrFile, []byte("content"), defaultFileMode)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			filePaths, err := config.GetCrYAMLFilePaths(tempDir)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(filePaths).To(gomega.ConsistOf(crFile1, crFile2))
			gomega.Expect(filePaths).NotTo(gomega.ContainElement(nonCrFile))
		})

		ginkgo.It("should return an error if the directory does not exist", func() {
			_, err := config.GetCrYAMLFilePaths("non-existent-directory")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})
