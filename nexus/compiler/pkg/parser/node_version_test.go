// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"

	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
)

var _ = Describe("Multi-version node parsing", func() {
	It("should merge a non-storage version struct into the storage version's OtherVersions", func() {
		dslPath := "../../example/test-utils/multi-version-datamodel"
		pkgs := parser.ParseDSLPkg(dslPath)
		graphqlQueries := parser.ParseGraphqlQuerySpecs(pkgs)
		graph, _, _ := parser.ParseDSLNodes(dslPath, baseGroupName, pkgs, graphqlQueries)

		root, ok := graph["roots.root."+baseGroupName]
		Expect(ok).To(BeTrue())

		org, ok := root.MultipleChildren["Org"]
		Expect(ok).To(BeTrue())

		Expect(org.Kind).To(Equal("Org"))
		Expect(org.Version).To(Equal("v1"))
		Expect(org.IsStorageVersion).To(BeTrue())

		Expect(org.OtherVersions).To(HaveLen(1))
		v2 := org.OtherVersions[0]
		Expect(v2.Name).To(Equal("OrgV2"))
		Expect(v2.Kind).To(Equal("Org"))
		Expect(v2.Version).To(Equal("v2"))
		Expect(v2.IsStorageVersion).To(BeFalse())
		Expect(v2.ConversionImplemented).To(BeTrue())

		// Both versions collapse onto the same CRD/Kind identity.
		Expect(v2.CrdName).To(Equal(org.CrdName))
	})

	It("should fail when a non-storage version does not implement ConvertTo/ConvertFrom", func() {
		defer func() { log.StandardLogger().ExitFunc = nil }()

		fail := false
		log.StandardLogger().ExitFunc = func(int) {
			fail = true
		}

		dslPath := "../../example/test-utils/multi-version-missing-conversion"
		pkgs := parser.ParseDSLPkg(dslPath)
		graphqlQueries := parser.ParseGraphqlQuerySpecs(pkgs)
		parser.ParseDSLNodes(dslPath, baseGroupName, pkgs, graphqlQueries)

		Expect(fail).To(BeTrue())
	})

	It("should fail when more than one version of a Kind claims to be the storage version", func() {
		defer func() { log.StandardLogger().ExitFunc = nil }()

		fail := false
		log.StandardLogger().ExitFunc = func(int) {
			fail = true
		}

		dslPath := "../../example/test-utils/multi-version-duplicate-storage"
		pkgs := parser.ParseDSLPkg(dslPath)
		graphqlQueries := parser.ParseGraphqlQuerySpecs(pkgs)
		parser.ParseDSLNodes(dslPath, baseGroupName, pkgs, graphqlQueries)

		Expect(fail).To(BeTrue())
	})
})
