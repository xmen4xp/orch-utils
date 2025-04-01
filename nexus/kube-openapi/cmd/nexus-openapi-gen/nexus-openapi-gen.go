// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// This package generates openAPI definition file to be used in open API spec generation on API servers. To generate
// definition for a specific type or package add "+k8s:openapi-gen=true" tag to the type/package comment lines. To
// exclude a type from a tagged package, add "+k8s:openapi-gen=false" tag to the type comment lines.

package main

import (
	"flag"
	"log"

	generatorargs "github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/cmd/nexus-openapi-gen/args"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/generators"

	"github.com/spf13/pflag"

	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	genericArgs, customArgs := generatorargs.NewDefaults()

	genericArgs.AddFlags(pflag.CommandLine)
	customArgs.AddFlags(pflag.CommandLine)
	flag.Set("logtostderr", "true")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if err := generatorargs.Validate(genericArgs); err != nil {
		log.Fatalf("Arguments validation error: %v", err)
	}

	// Generates the code for the OpenAPIDefinitions.
	if err := genericArgs.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		log.Fatalf("OpenAPI code generation error: %v", err)
	}
}
