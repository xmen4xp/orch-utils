// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/codegen/config"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/plugin"
)

type Option func(cfg *config.Config, plugins *[]plugin.Plugin)

func NoPlugins() Option {
	return func(cfg *config.Config, plugins *[]plugin.Plugin) {
		*plugins = nil
	}
}

func AddPlugin(p plugin.Plugin) Option {
	return func(cfg *config.Config, plugins *[]plugin.Plugin) {
		*plugins = append(*plugins, p)
	}
}

// PrependPlugin prepends plugin any existing plugins
func PrependPlugin(p plugin.Plugin) Option {
	return func(cfg *config.Config, plugins *[]plugin.Plugin) {
		*plugins = append([]plugin.Plugin{p}, *plugins...)
	}
}

// ReplacePlugin replaces any existing plugin with a matching plugin name
func ReplacePlugin(p plugin.Plugin) Option {
	return func(cfg *config.Config, plugins *[]plugin.Plugin) {
		if plugins != nil {
			found := false
			ps := *plugins
			for i, o := range ps {
				if p.Name() == o.Name() {
					ps[i] = p
					found = true
				}
			}
			if !found {
				ps = append(ps, p)
			}
			*plugins = ps
		}
	}
}
