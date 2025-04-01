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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/open-edge-platform/orch-utils/tenancy-api-mapping/pkg/config"
	"github.com/open-edge-platform/orch-utils/tenancy-api-mapping/pkg/git"
	"github.com/open-edge-platform/orch-utils/tenancy-api-mapping/pkg/openapi"
)

func main() {
	if err := Run("config.yaml"); err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Println("\nSpecGen Completed!")
}

func Run(configPath string) error {
	// read config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// read APIMappingConfig CRs
	filePaths, err := config.GetCrYAMLFilePaths(cfg.Global.APImappingConfigCrsDir)
	if err != nil {
		return fmt.Errorf("failed to read filepaths from the directory = %s: %w", cfg.Global.APImappingConfigCrsDir, err)
	}

	// Create an instance of ExecCmdRunner
	execRunner := &git.ExecCmdRunner{}
	for _, filePath := range filePaths {
		fmt.Println("about to start processing for filePath:", filePath)
		fBytes, err := config.ReadCRFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read contents from file = %s: %w", filePath, err)
		}

		// Parse the YAML content to get the RepoConf object
		mappingConfigCr, err := config.ParseCRContent(fBytes)
		if err != nil {
			return fmt.Errorf("error parsing cr content for file= %s: %w", filePath, err)
		}

		if !mappingConfigCr.Spec.SpecGenEnabled {
			fmt.Println("skipping generating spec for :", mappingConfigCr.Spec.RepoConf.URL, "as SpecGenEnabled is false")
			continue
		}

		repoConf := mappingConfigCr.Spec.RepoConf
		err = git.InitSubmodule(execRunner, cfg.Global.LocalSubModsDir, repoConf.URL, repoConf.Tag, mappingConfigCr.Metadata.Name)
		if err != nil {
			return fmt.Errorf("failed to initialize submodule: %w", err)
		}

		err = openapi.ProcessOpenAPISpec(mappingConfigCr, cfg.Global)
		if err != nil {
			return fmt.Errorf("failed to create multi-tenancy OpenAPI spec for mapping conf %s : %w",
				mappingConfigCr.Metadata.Name, err)
		}

		fmt.Printf("Created Multi-tenancy OpenAPI spec for repo: %s\n", repoConf.URL)

		// cleanup
		os.RemoveAll(cfg.Global.LocalSubModsDir)
	}

	return nil
}
