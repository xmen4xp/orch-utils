// Copyright 2016 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"sort"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/common"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
)

type parameters []spec.Parameter

func (s parameters) Len() int      { return len(s) }
func (s parameters) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// byNameIn used in sorting parameters by Name and In fields.
type byNameIn struct {
	parameters
}

func (s byNameIn) Less(i, j int) bool {
	return s.parameters[i].Name < s.parameters[j].Name || (s.parameters[i].Name == s.parameters[j].Name && s.parameters[i].In < s.parameters[j].In)
}

// SortParameters sorts parameters by Name and In fields.
func sortParameters(p []spec.Parameter) {
	sort.Sort(byNameIn{p})
}

func groupRoutesByPath(routes []common.Route) map[string][]common.Route {
	pathToRoutes := make(map[string][]common.Route)
	for _, r := range routes {
		pathToRoutes[r.Path()] = append(pathToRoutes[r.Path()], r)
	}
	return pathToRoutes
}

func mapKeyFromParam(param common.Parameter) interface{} {
	return struct {
		Name string
		Kind common.ParameterKind
	}{
		Name: param.Name(),
		Kind: param.Kind(),
	}
}
