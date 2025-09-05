package controllers

import (
	"api-gw/pkg/model"
	"context"
	"encoding/json"

	tenantv1 "nexus/admin/api/build/apis/tenantconfig.admin.nexus.com/v1"

	yamlv1 "github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TenantConfigController controller", func() {
	It("should process tenant config", func() {
		crdJson, err := yamlv1.YAMLToJSON([]byte(tenantConfigExample))
		Expect(err).NotTo(HaveOccurred())

		var obj tenantv1.Tenant
		err = json.Unmarshal(crdJson, &obj)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Create(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())

		event := <-model.TenantEvent
		Expect(event.Type).To(Equal(model.Upsert))
		Expect(event.Tenant.Name).To(Equal("tenant1"))

		err = k8sClient.Delete(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())

		event = <-model.TenantEvent
		Expect(event.Type).To(Equal(model.Delete))
	})
})
