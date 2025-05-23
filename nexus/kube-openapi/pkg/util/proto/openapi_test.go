// Copyright 2017 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proto_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util/proto"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util/proto/testing"
)

var fakeSchema = testing.Fake{Path: filepath.Join("testdata", "swagger.json")}
var fakeSchemaNext = testing.Fake{Path: filepath.Join("testdata", "swagger_next.json")}
var fakeSchemaV300 = testing.FakeV3{Path: filepath.Join("testdata", "openapi_v3_0_0")}

var _ = Describe("Reading apps/v1beta1/Deployment from v1.8 openAPIData", func() {
	var models proto.Models
	BeforeEach(func() {
		s, err := fakeSchema.OpenAPISchema()
		Expect(err).To(BeNil())
		models, err = proto.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	model := "io.k8s.api.apps.v1beta1.Deployment"
	var schema proto.Schema
	It("should lookup the Schema by its model name", func() {
		schema = models.LookupModel(model)
		Expect(schema).ToNot(BeNil())
	})

	var deployment *proto.Kind
	It("should be a Kind", func() {
		deployment = schema.(*proto.Kind)
		Expect(deployment).ToNot(BeNil())
	})

	It("should have a path", func() {
		Expect(deployment.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.Deployment"}))
	})

	It("should have a kind key of type string", func() {
		Expect(deployment.Fields).To(HaveKey("kind"))
		key := deployment.Fields["kind"].(*proto.Primitive)
		Expect(key).ToNot(BeNil())
		Expect(key.Type).To(Equal("string"))
		Expect(key.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.Deployment", ".kind"}))
	})

	It("should have a apiVersion key of type string", func() {
		Expect(deployment.Fields).To(HaveKey("apiVersion"))
		key := deployment.Fields["apiVersion"].(*proto.Primitive)
		Expect(key).ToNot(BeNil())
		Expect(key.Type).To(Equal("string"))
		Expect(key.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.Deployment", ".apiVersion"}))
	})

	It("should have a metadata key of type Reference", func() {
		Expect(deployment.Fields).To(HaveKey("metadata"))
		key := deployment.Fields["metadata"].(proto.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"))
		subSchema := key.SubSchema().(*proto.Kind)
		Expect(subSchema).ToNot(BeNil())
	})

	var status *proto.Kind
	It("should have a status key of type Reference", func() {
		Expect(deployment.Fields).To(HaveKey("status"))
		key := deployment.Fields["status"].(proto.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.api.apps.v1beta1.DeploymentStatus"))
		status = key.SubSchema().(*proto.Kind)
		Expect(status).ToNot(BeNil())
	})

	It("should have a valid DeploymentStatus", func() {
		By("having availableReplicas key")
		Expect(status.Fields).To(HaveKey("availableReplicas"))
		replicas := status.Fields["availableReplicas"].(*proto.Primitive)
		Expect(replicas).ToNot(BeNil())
		Expect(replicas.Type).To(Equal("integer"))

		By("having conditions key")
		Expect(status.Fields).To(HaveKey("conditions"))
		conditions := status.Fields["conditions"].(*proto.Array)
		Expect(conditions).ToNot(BeNil())
		Expect(conditions.GetName()).To(Equal(`Array of Reference to "io.k8s.api.apps.v1beta1.DeploymentCondition"`))
		Expect(conditions.GetExtensions()).To(Equal(map[string]interface{}{
			"x-kubernetes-patch-merge-key": "type",
			"x-kubernetes-patch-strategy":  "merge",
		}))
		condition := conditions.SubType.(proto.Reference)
		Expect(condition.Reference()).To(Equal("io.k8s.api.apps.v1beta1.DeploymentCondition"))
	})

	var spec *proto.Kind
	It("should have a spec key of type Reference", func() {
		Expect(deployment.Fields).To(HaveKey("spec"))
		key := deployment.Fields["spec"].(proto.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.api.apps.v1beta1.DeploymentSpec"))
		spec = key.SubSchema().(*proto.Kind)
		Expect(spec).ToNot(BeNil())
	})

	It("should have a spec with no gvk", func() {
		_, found := spec.GetExtensions()["x-kubernetes-group-version-kind"]
		Expect(found).To(BeFalse())
	})

	It("should have a spec with a PodTemplateSpec sub-field", func() {
		Expect(spec.Fields).To(HaveKey("template"))
		key := spec.Fields["template"].(proto.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.api.core.v1.PodTemplateSpec"))
	})
})

var _ = Describe("Reading apps/v1beta1/Deployment from v1.11 openAPIData", func() {
	var models proto.Models
	BeforeEach(func() {
		s, err := fakeSchemaNext.OpenAPISchema()
		Expect(err).To(BeNil())
		models, err = proto.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	model := "io.k8s.api.apps.v1beta1.Deployment"
	var schema proto.Schema
	It("should lookup the Schema by its model name", func() {
		schema = models.LookupModel(model)
		Expect(schema).ToNot(BeNil())
	})

	var deployment *proto.Kind
	It("should be a Kind", func() {
		deployment = schema.(*proto.Kind)
		Expect(deployment).ToNot(BeNil())
	})
})

var _ = Describe("Reading apps/v1beta1/ControllerRevision from v1.11 openAPIData", func() {
	var models proto.Models
	BeforeEach(func() {
		s, err := fakeSchemaNext.OpenAPISchema()
		Expect(err).To(BeNil())
		models, err = proto.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	model := "io.k8s.api.apps.v1beta1.ControllerRevision"
	var schema proto.Schema
	It("should lookup the Schema by its model name", func() {
		schema = models.LookupModel(model)
		Expect(schema).ToNot(BeNil())
	})

	var cr *proto.Kind
	It("data property should be map[string]Arbitrary", func() {
		cr = schema.(*proto.Kind)
		Expect(cr).ToNot(BeNil())
		Expect(cr.Fields).To(HaveKey("data"))

		data := cr.Fields["data"].(*proto.Map)
		Expect(data).ToNot(BeNil())
		Expect(data.GetName()).To(Equal("Map of Arbitrary value (primitive, object or array)"))
		Expect(data.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.ControllerRevision", ".data"}))

		arbitrary := data.SubType.(*proto.Arbitrary)
		Expect(arbitrary).ToNot(BeNil())
		Expect(arbitrary.GetName()).To(Equal("Arbitrary value (primitive, object or array)"))
		Expect(arbitrary.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.ControllerRevision", ".data"}))
	})
})

var _ = Describe("Reading authorization.k8s.io/v1/SubjectAccessReview from openAPIData", func() {
	var models proto.Models
	BeforeEach(func() {
		s, err := fakeSchema.OpenAPISchema()
		Expect(err).To(BeNil())
		models, err = proto.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	model := "io.k8s.api.authorization.v1.LocalSubjectAccessReview"
	var schema proto.Schema
	It("should lookup the Schema by its model", func() {
		schema = models.LookupModel(model)
		Expect(schema).ToNot(BeNil())
	})

	var sarspec *proto.Kind
	It("should be a Kind and have a spec", func() {
		sar := schema.(*proto.Kind)
		Expect(sar).ToNot(BeNil())
		Expect(sar.Fields).To(HaveKey("spec"))
		specRef := sar.Fields["spec"].(proto.Reference)
		Expect(specRef).ToNot(BeNil())
		Expect(specRef.Reference()).To(Equal("io.k8s.api.authorization.v1.SubjectAccessReviewSpec"))
		sarspec = specRef.SubSchema().(*proto.Kind)
		Expect(sarspec).ToNot(BeNil())
	})

	It("should have a valid SubjectAccessReviewSpec", func() {
		Expect(sarspec.Fields).To(HaveKey("extra"))
		extra := sarspec.Fields["extra"].(*proto.Map)
		Expect(extra).ToNot(BeNil())
		Expect(extra.GetName()).To(Equal("Map of Array of string"))
		Expect(extra.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
		array := extra.SubType.(*proto.Array)
		Expect(array).ToNot(BeNil())
		Expect(array.GetName()).To(Equal("Array of string"))
		Expect(array.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
		str := array.SubType.(*proto.Primitive)
		Expect(str).ToNot(BeNil())
		Expect(str.Type).To(Equal("string"))
		Expect(str.GetName()).To(Equal("string"))
		Expect(str.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
	})
})

var _ = Describe("Path", func() {
	It("can be created by NewPath", func() {
		path := proto.NewPath("key")
		Expect(path.String()).To(Equal("key"))
	})
	It("can create and print complex paths", func() {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.String()).To(Equal("key[12].subKey"))
	})
	It("has a length", func() {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.Len()).To(Equal(3))
	})
	It("can look like an array", func() {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.Get()).To(Equal([]string{"key", "[12]", ".subKey"}))
	})
})

var _ = Describe("Reading apps/v1/Deployment from v3.0.0 openAPIData", func() {
	var deployment *proto.Kind
	BeforeEach(func() {
		var models proto.Models
		s, schemaErr := fakeSchemaV300.OpenAPIV3Schema("apps/v1")
		models, modelsErr := proto.NewOpenAPIV3Data(s)

		Expect(schemaErr).To(BeNil())
		Expect(modelsErr).To(BeNil())

		model := "io.k8s.api.apps.v1.Deployment"
		schema := models.LookupModel(model)
		Expect(schema).ToNot(BeNil())

		deployment = schema.(*proto.Kind)
		Expect(deployment).ToNot(BeNil())
	})

	It("should have a path", func() {
		Expect(deployment.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1.Deployment"}))
	})

	It("should have a kind key of type string", func() {
		Expect(deployment.Fields).To(HaveKey("kind"))
		key := deployment.Fields["kind"].(*proto.Primitive)
		Expect(key).ToNot(BeNil())
		Expect(key.Type).To(Equal("string"))
		Expect(key.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1.Deployment", ".kind"}))
	})

	It("should have a apiVersion key of type string", func() {
		Expect(deployment.Fields).To(HaveKey("apiVersion"))
		key := deployment.Fields["apiVersion"].(*proto.Primitive)
		Expect(key).ToNot(BeNil())
		Expect(key.Type).To(Equal("string"))
		Expect(key.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1.Deployment", ".apiVersion"}))
	})

	It("should have a metadata key of type Reference", func() {
		Expect(deployment.Fields).To(HaveKey("metadata"))
		key := deployment.Fields["metadata"].(proto.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"))
		subSchema := key.SubSchema().(*proto.Kind)
		Expect(subSchema).ToNot(BeNil())
	})

	Describe("status", func() {
		var status *proto.Kind
		BeforeEach(func() {
			Expect(deployment.Fields).To(HaveKey("status"))
			key := deployment.Fields["status"].(proto.Reference)
			Expect(key).ToNot(BeNil())
			Expect(key.Reference()).To(Equal("io.k8s.api.apps.v1.DeploymentStatus"))
			status = key.SubSchema().(*proto.Kind)
			Expect(status).ToNot(BeNil())
		})

		It("should have a valid DeploymentStatus", func() {
			By("having availableReplicas key")
			Expect(status.Fields).To(HaveKey("availableReplicas"))
			replicas := status.Fields["availableReplicas"].(*proto.Primitive)
			Expect(replicas).ToNot(BeNil())
			Expect(replicas.Type).To(Equal("integer"))

			By("having conditions key")
			Expect(status.Fields).To(HaveKey("conditions"))
			conditions := status.Fields["conditions"].(*proto.Array)
			Expect(conditions).ToNot(BeNil())
			Expect(conditions.GetName()).To(Equal(`Array of Reference to "io.k8s.api.apps.v1.DeploymentCondition"`))
			Expect(conditions.GetExtensions()).To(Equal(map[string]interface{}{
				"x-kubernetes-patch-merge-key": "type",
				"x-kubernetes-patch-strategy":  "merge",
			}))
			condition := conditions.SubType.(proto.Reference)
			Expect(condition.Reference()).To(Equal("io.k8s.api.apps.v1.DeploymentCondition"))
		})
	})

	Describe("spec subschema", func() {
		var spec *proto.Kind
		BeforeEach(func() {
			Expect(deployment.Fields).To(HaveKey("spec"))
			key, _ := deployment.Fields["spec"].(proto.Reference)
			Expect(key).ToNot(BeNil())
			Expect(key.Reference()).To(Equal("io.k8s.api.apps.v1.DeploymentSpec"))
			spec = key.SubSchema().(*proto.Kind)
			Expect(spec).ToNot(BeNil())
		})

		It("should have a spec with no gvk", func() {
			_, found := spec.GetExtensions()["x-kubernetes-group-version-kind"]
			Expect(found).To(BeFalse())
		})

		It("should have a spec with a PodTemplateSpec sub-field", func() {
			Expect(spec.Fields).To(HaveKey("template"))
			key := spec.Fields["template"].(proto.Reference)
			Expect(key).ToNot(BeNil())
			Expect(key.Reference()).To(Equal("io.k8s.api.core.v1.PodTemplateSpec"))
		})
	})
})
