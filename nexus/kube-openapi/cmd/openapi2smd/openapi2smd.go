// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io/ioutil"
	"log"
	"os"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	yaml "gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/schemaconv"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util/proto"
)

func main() {
	if len(os.Args) != 1 {
		log.Fatal("this program takes input on stdin and writes output to stdout.")
	}

	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error reading stdin: %v", err)
	}

	document, err := openapi_v2.ParseDocument(input)
	if err != nil {
		log.Fatalf("error interpreting stdin: %v", err)
	}

	models, err := proto.NewOpenAPIData(document)
	if err != nil {
		log.Fatalf("error interpreting models: %v", err)
	}

	newSchema, err := schemaconv.ToSchema(models)
	if err != nil {
		log.Fatalf("error converting schema format: %v", err)
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(newSchema); err != nil {
		log.Fatalf("error writing new schema: %v", err)
	}

}
