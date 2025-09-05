package controllers

import (
	"context"
	"encoding/json"

	apiv1 "nexus/admin/api/build/apis/api.admin.nexus.com/v1"
	configv1 "nexus/admin/api/build/apis/config.admin.nexus.com/v1"
	routev1 "nexus/admin/api/build/apis/route.admin.nexus.com/v1"

	yamlv1 "github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route controller", func() {
	It("should create nexus obj", func() {
		crdJson, err := yamlv1.YAMLToJSON([]byte(nexusExample))
		Expect(err).NotTo(HaveOccurred())

		var obj apiv1.Nexus
		err = json.Unmarshal(crdJson, &obj)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Create(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())
	})
	It("should create config obj", func() {
		crdJson, err := yamlv1.YAMLToJSON([]byte(nexusExample))
		Expect(err).NotTo(HaveOccurred())

		var obj configv1.Config
		err = json.Unmarshal(crdJson, &obj)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Create(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())
	})
	It("should create a route", func() {
		crdJson, err := yamlv1.YAMLToJSON([]byte(routeExample))
		Expect(err).NotTo(HaveOccurred())

		var obj routev1.Route
		err = json.Unmarshal(crdJson, &obj)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Create(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Delete(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())
	})
})
