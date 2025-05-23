// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	headerFilePath = "../../boilerplate/boilerplate.go.txt"
	testdataDir    = "./testdata"
	testPkgDir     = "github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/test/integration/testdata"
	inputDir       = testPkgDir + "/listtype" +
		"," + testPkgDir + "/maptype" +
		"," + testPkgDir + "/structtype" +
		"," + testPkgDir + "/dummytype" +
		"," + testPkgDir + "/uniontype" +
		"," + testPkgDir + "/enumtype" +
		"," + testPkgDir + "/custom" +
		"," + testPkgDir + "/defaults"
	outputBase                 = "pkg"
	outputPackage              = "generated"
	outputBaseFileName         = "openapi_generated"
	generatedSwaggerFileName   = "generated.v2.json"
	generatedReportFileName    = "generated.v2.report"
	goldenSwaggerFileName      = "golden.v2.json"
	goldenReportFileName       = "golden.v2.report"
	generatedOpenAPIv3FileName = "generated.v3.json"
	goldenOpenAPIv3Filename    = "golden.v3.json"

	timeoutSeconds = 10.0
)

func TestGenerators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var _ = Describe("Open API Definitions Generation", func() {

	var (
		workingDirectory string
		tempDir          string
		terr             error
		openAPIGenPath   string
	)

	testdataFile := func(filename string) string { return filepath.Join(testdataDir, filename) }
	generatedFile := func(filename string) string { return filepath.Join(tempDir, filename) }

	BeforeSuite(func() {
		// Explicitly manage working directory
		abs, err := filepath.Abs("")
		Expect(err).ShouldNot(HaveOccurred())
		workingDirectory = abs

		// Create a temporary directory for generated swagger files.
		tempDir, terr = ioutil.TempDir("./", "openapi")
		Expect(terr).ShouldNot(HaveOccurred())

		// Build the OpenAPI code generator.
		By("building openapi-gen")
		binaryPath, berr := gexec.Build("../../cmd/openapi-gen/openapi-gen.go")
		Expect(berr).ShouldNot(HaveOccurred())
		openAPIGenPath = binaryPath

		// Run the OpenAPI code generator, creating OpenAPIDefinition code
		// to be compiled into builder.
		By("processing go idl with openapi-gen")
		gr := generatedFile(generatedReportFileName)
		command := exec.Command(openAPIGenPath,
			"-i", inputDir,
			"-o", outputBase,
			"-p", outputPackage,
			"-O", outputBaseFileName,
			"-r", gr,
			"-h", headerFilePath,
		)
		command.Dir = workingDirectory
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, timeoutSeconds).Should(gexec.Exit(0))

		By("writing swagger v2.0")
		// Create the OpenAPI swagger builder.
		binaryPath, berr = gexec.Build("./builder/main.go")
		Expect(berr).ShouldNot(HaveOccurred())

		// Execute the builder, generating an OpenAPI swagger file with definitions.
		gs := generatedFile(generatedSwaggerFileName)
		By("writing swagger to " + gs)
		command = exec.Command(binaryPath, gs)
		command.Dir = workingDirectory
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, timeoutSeconds).Should(gexec.Exit(0))

		By("writing OpenAPI v3.0")
		// Create the OpenAPI swagger builder.
		binaryPath, berr = gexec.Build("./builder3/main.go")
		Expect(berr).ShouldNot(HaveOccurred())

		// Execute the builder, generating an OpenAPI swagger file with definitions.
		gov3 := generatedFile(generatedOpenAPIv3FileName)
		By("writing swagger to " + gov3)
		command = exec.Command(binaryPath, gov3)
		command.Dir = workingDirectory
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
	})

	AfterSuite(func() {
		os.RemoveAll(tempDir)
		gexec.CleanupBuildArtifacts()
	})

	Describe("openapi-gen --verify", func() {
		It("Verifies that the existing files are correct", func() {
			command := exec.Command(openAPIGenPath,
				"-i", inputDir,
				"-o", outputBase,
				"-p", outputPackage,
				"-O", outputBaseFileName,
				"-r", testdataFile(goldenReportFileName),
				"-h", headerFilePath,
				"--verify-only",
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})

	Describe("Validating OpenAPI V2 Definition Generation", func() {
		It("Generated OpenAPI swagger definitions should match golden files", func() {
			// Diff the generated swagger against the golden swagger. Exit code should be zero.
			command := exec.Command(
				"diff",
				testdataFile(goldenSwaggerFileName),
				generatedFile(generatedSwaggerFileName),
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})

	Describe("Validating OpenAPI V3 Definition Generation", func() {
		It("Generated OpenAPI swagger definitions should match golden files", func() {
			// Diff the generated swagger against the golden swagger. Exit code should be zero.
			command := exec.Command(
				"diff",
				testdataFile(goldenOpenAPIv3Filename),
				generatedFile(generatedOpenAPIv3FileName),
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})

	Describe("Validating API Rule Violation Reporting", func() {
		It("Generated API rule violations should match golden report files", func() {
			// Diff the generated report against the golden report. Exit code should be zero.
			command := exec.Command(
				"diff",
				testdataFile(goldenReportFileName),
				generatedFile(generatedReportFileName),
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})
})
