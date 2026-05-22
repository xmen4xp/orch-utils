package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/config"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
	nexus "github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

var _ = Describe("ExtensionRestAPI parsing tests", func() {
	Describe("ValidateOpenAPIPathSpec", func() {
		It("should pass validation with valid OpenAPI path spec", func() {
			spec := `
get:
  summary: Get resource
  responses:
    "200":
      description: OK
`
			err := parser.ValidateOpenAPIPathSpec(spec)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should pass validation with multiple HTTP methods", func() {
			spec := `
get:
  summary: Get resource
  responses:
    "200":
      description: OK
post:
  summary: Create resource
  responses:
    "201":
      description: Created
`
			err := parser.ValidateOpenAPIPathSpec(spec)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail validation with invalid YAML", func() {
			spec := `
get: [invalid: yaml: content
`
			err := parser.ValidateOpenAPIPathSpec(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid YAML syntax"))
		})

		It("should fail validation with no HTTP methods", func() {
			spec := `
parameters:
  - name: id
    in: path
`
			err := parser.ValidateOpenAPIPathSpec(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must define at least one HTTP method"))
		})
	})

	Describe("ValidateExtensionRestAPI", func() {
		It("should pass validation with valid spec", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:    "myApi",
				PkgName: "mypkg",
				Uri:     "/v1/resources/{id}",
				OpenAPIPathSpec: `
get:
  summary: Get resource
  responses:
    "200":
      description: OK
`,
			}
			err := parser.ValidateExtensionRestAPI(spec)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should pass validation without OpenAPIPathSpec (optional)", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:    "myApi",
				PkgName: "mypkg",
				Uri:     "/v1/resources/{id}",
			}
			err := parser.ValidateExtensionRestAPI(spec)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail validation without uri", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:    "myApi",
				PkgName: "mypkg",
			}
			err := parser.ValidateExtensionRestAPI(spec)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("uri is required"))
		})

		It("should fail validation with invalid OpenAPIPathSpec", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:            "myApi",
				PkgName:         "mypkg",
				Uri:             "/v1/resources/{id}",
				OpenAPIPathSpec: `invalid: yaml: content`,
			}
			err := parser.ValidateExtensionRestAPI(spec)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ValidateExtensionRestAPIPathParams", func() {
		var parentsMap map[string]parser.NodeHelper

		BeforeEach(func() {
			// Set up a mock parentsMap with a hierarchy:
			// root.Root (singleton) -> config.Config -> gns.Gns
			parentsMap = map[string]parser.NodeHelper{
				"roots.root.test.com": {
					Name:        "Root",
					RestName:    "root.Root",
					Parents:     []string{},
					IsSingleton: true,
				},
				"configs.config.test.com": {
					Name:     "Config",
					RestName: "config.Config",
					Parents:  []string{"roots.root.test.com"},
				},
				"gnses.gns.test.com": {
					Name:     "Gns",
					RestName: "gns.Gns",
					Parents:  []string{"roots.root.test.com", "configs.config.test.com"},
				},
			}
		})

		It("should pass validation with valid path params in hierarchy", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:           "myApi",
				PkgName:        "gns",
				Uri:            "/v1/config/{config.Config}/gns/{gns.Gns}/custom",
				AssociatedNode: "gns.Gns",
				NodeCRDName:    "gnses.gns.test.com",
			}
			err := parser.ValidateExtensionRestAPIPathParams(spec, parentsMap)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail validation with path param not in hierarchy", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:           "myApi",
				PkgName:        "gns",
				Uri:            "/v1/foo/{foo.Bar}/gns/{gns.Gns}",
				AssociatedNode: "gns.Gns",
				NodeCRDName:    "gnses.gns.test.com",
			}
			err := parser.ValidateExtensionRestAPIPathParams(spec, parentsMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("path param {foo.Bar} not found in hierarchy"))
		})

		It("should fail validation without associated node", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:    "myApi",
				PkgName: "gns",
				Uri:     "/v1/gns/{gns.Gns}",
			}
			err := parser.ValidateExtensionRestAPIPathParams(spec, parentsMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not associated with any node"))
		})

		It("should fail validation when required non-singleton parent is missing from URI", func() {
			spec := parser.ExtensionRestAPISpec{
				Name:           "myApi",
				PkgName:        "gns",
				Uri:            "/v1/gns/{gns.Gns}/custom",
				AssociatedNode: "gns.Gns",
				NodeCRDName:    "gnses.gns.test.com",
			}
			err := parser.ValidateExtensionRestAPIPathParams(spec, parentsMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("required parent path params missing from URI"))
			Expect(err.Error()).To(ContainSubstring("config.Config"))
		})

		It("should pass validation when singleton parent is missing from URI", func() {
			// root.Root is singleton, so it's allowed to be absent
			spec := parser.ExtensionRestAPISpec{
				Name:           "myApi",
				PkgName:        "config",
				Uri:            "/v1/config/{config.Config}/custom",
				AssociatedNode: "config.Config",
				NodeCRDName:    "configs.config.test.com",
			}
			err := parser.ValidateExtensionRestAPIPathParams(spec, parentsMap)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should pass validation when missing parent is in ignoredParentPathParams", func() {
			// Temporarily add config.Config to ignored list
			origIgnored := config.ConfigInstance.IgnoredParentPathParams
			config.ConfigInstance.IgnoredParentPathParams = []string{"config.Config"}
			defer func() { config.ConfigInstance.IgnoredParentPathParams = origIgnored }()

			spec := parser.ExtensionRestAPISpec{
				Name:           "myApi",
				PkgName:        "gns",
				Uri:            "/v1/gns/{gns.Gns}/custom",
				AssociatedNode: "gns.Gns",
				NodeCRDName:    "gnses.gns.test.com",
			}
			err := parser.ValidateExtensionRestAPIPathParams(spec, parentsMap)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ParseOpenAPIPathSpecToRestAPISpec", func() {
		It("should parse valid OpenAPI path spec with GET method", func() {
			openAPISpec := `
get:
  summary: Get resource
  responses:
    "200":
      description: OK
    "404":
      description: Not Found
`
			restAPISpec, err := parser.ParseOpenAPIPathSpecToRestAPISpec("/v1/resources", openAPISpec)
			Expect(err).NotTo(HaveOccurred())
			Expect(restAPISpec.Uris).To(HaveLen(1))
			Expect(restAPISpec.Uris[0].Uri).To(Equal("/v1/resources"))
			Expect(restAPISpec.Uris[0].Methods).To(HaveKey(nexus.HTTPMethod("GET")))
		})

		It("should parse multiple HTTP methods", func() {
			openAPISpec := `
get:
  responses:
    "200":
      description: OK
put:
  responses:
    "200":
      description: Updated
delete:
  responses:
    "204":
      description: No Content
`
			restAPISpec, err := parser.ParseOpenAPIPathSpecToRestAPISpec("/v1/resources", openAPISpec)
			Expect(err).NotTo(HaveOccurred())
			Expect(restAPISpec.Uris[0].Methods).To(HaveKey(nexus.HTTPMethod("GET")))
			Expect(restAPISpec.Uris[0].Methods).To(HaveKey(nexus.HTTPMethod("PUT")))
			Expect(restAPISpec.Uris[0].Methods).To(HaveKey(nexus.HTTPMethod("DELETE")))
		})

		It("should handle empty OpenAPI spec", func() {
			restAPISpec, err := parser.ParseOpenAPIPathSpecToRestAPISpec("/v1/resources", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(restAPISpec.Uris).To(HaveLen(1))
			Expect(restAPISpec.Uris[0].Uri).To(Equal("/v1/resources"))
		})
	})
})
