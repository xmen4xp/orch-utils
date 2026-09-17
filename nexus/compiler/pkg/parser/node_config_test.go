// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
)

var _ = Describe("Node config tests", func() {
	var (
		//err error
		pkg parser.Package
		ok  bool
	)

	BeforeEach(func() {
		pkgs := parser.ParseDSLPkg(exampleDSLPath)
		pkg, ok = pkgs["github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/datamodel/config/gns"]
		Expect(ok).To(BeTrue())
	})

	It("should parse gns node annotation", func() {
		annotation, ok := parser.GetNexusRestAPIGenAnnotation(pkg, "Gns")
		Expect(ok).To(BeTrue())
		Expect(annotation).To(Equal("GNSRestAPISpec"))
	})

	It("should parse the nexus-on-delete child tag", func() {
		policyPkgs := parser.ParseDSLPkg("../../example/test-utils/deletion-policy-datamodel")
		policyPkg, found := policyPkgs["example.com/deletion-policy-datamodel"]
		Expect(found).To(BeTrue())

		var rootFound bool
		var restrictCount int
		for _, n := range policyPkg.GetNexusNodes() {
			if parser.GetTypeName(n) != "Root" {
				continue
			}
			rootFound = true
			for _, f := range parser.GetChildFields(n) {
				if policy, present := parser.GetChildOnDeletePolicy(f); present {
					Expect(policy).To(Equal(parser.DeletionPolicyRestrict))
					restrictCount++
				}
			}
		}
		Expect(rootFound).To(BeTrue())
		// Exactly one child (AISlice) is tagged restrict; Foo has no tag.
		Expect(restrictCount).To(Equal(1))
	})
})
