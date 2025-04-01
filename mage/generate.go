// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package mage

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/bitfield/script"
)

func jwtPluginConfigmap() error {
	const (
		outputDir = "./charts/traefik-pre/templates/"
		fileName  = "jwt-plugin-configmap-generated.yaml"
		sourceDir = "./traefik-plugins/jwt-plugin/vendor/github.com/team-carepay/traefik-jwt-plugin/"
	)

	kubeCmd := fmt.Sprintf("kubectl create configmap jwt-plugin --from-file=%s --dry-run=client -n orch-gateway -o yaml", sourceDir) //nolint: lll

	stderr := &bytes.Buffer{}
	_, err := script.NewPipe().Exec(kubeCmd).WithStderr(stderr).WriteFile(filepath.Join(outputDir, fileName))
	if err != nil {
		return fmt.Errorf("generating jwt plugin configmap: %w: %s", err, stderr)
	}
	return nil
}
