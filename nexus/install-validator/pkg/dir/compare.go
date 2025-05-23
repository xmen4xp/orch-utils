// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package dir

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	nexuscompare "github.com/vmware-tanzu/graph-framework-for-microservices/common-library/pkg/nexus-compare"
	kubewrapper "github.com/vmware-tanzu/graph-framework-for-microservices/install-validator/pkg/k8s-utils"
	"sigs.k8s.io/yaml"
)

type compareFunc func([]byte, []byte) (bool, *bytes.Buffer, error)

func CheckDir(dir string, c kubewrapper.ClientInt, cFunc compareFunc) (map[string]*bytes.Buffer, []string, error) {
	changes := make(map[string]*bytes.Buffer)
	var (
		toDelete  []string
		toInstall []string
	)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			logrus.Debugf("%s is a directory, skipping", info.Name())
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".yaml") {
			logrus.Debugf("%s is not a yaml file, skipping", info.Name())
			return nil
		}
		newData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		name, err := nexuscompare.GetSpecName(newData)
		if err != nil {
			return err
		}
		crd := c.GetCrd(name)
		toInstall = append(toInstall, name)
		if crd == nil { //the crd is New, so no incompatible changes
			logrus.Debugf("%s not found in cluster, skipping comparison", name)
			return nil
		}

		if cFunc == nil {
			return errors.New("nil compare func passed")
		}
		actData, err := yaml.Marshal(crd)
		if err != nil {
			return err
		}

		hasAnyIncChanges, incChangesBuffer, err := cFunc(actData, newData)
		if err != nil {
			return err
		}

		if hasAnyIncChanges {
			changes[name] = incChangesBuffer
		}
		return nil
	})

	if err != nil {
		return changes, []string{}, err
	}

	for _, crd := range c.GetCrds() {
		found := false
		for _, nameToInstall := range toInstall {
			if crd.Name == nameToInstall {
				found = true
				break
			}
		}
		if !found && strings.HasSuffix(crd.Spec.Group, c.GetGroup()) {
			toDelete = append(toDelete, crd.Name)
		}
	}

	return changes, toDelete, err
}
