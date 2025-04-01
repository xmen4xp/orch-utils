// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package rules

import (
	"reflect"
	"strings"

	"k8s.io/gengo/types"
)

// OmitEmptyMatchCase implements APIRule interface.
// "omitempty" must appear verbatim (no case variants).
type OmitEmptyMatchCase struct{}

func (n *OmitEmptyMatchCase) Name() string {
	return "omitempty_match_case"
}

func (n *OmitEmptyMatchCase) Validate(t *types.Type) ([]string, error) {
	fields := make([]string, 0)

	// Only validate struct type and ignore the rest
	switch t.Kind {
	case types.Struct:
		for _, m := range t.Members {
			goName := m.Name
			jsonTag, ok := reflect.StructTag(m.Tags).Lookup("json")
			if !ok {
				continue
			}

			parts := strings.Split(jsonTag, ",")
			if len(parts) < 2 {
				// no tags other than name
				continue
			}
			if parts[0] == "-" {
				// not serialized
				continue
			}
			for _, part := range parts[1:] {
				if strings.EqualFold(part, "omitempty") && part != "omitempty" {
					fields = append(fields, goName)
				}
			}
		}
	}
	return fields, nil
}
