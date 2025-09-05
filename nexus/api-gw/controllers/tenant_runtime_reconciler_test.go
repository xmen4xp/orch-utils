package controllers

import (
	"api-gw/pkg/common"
	"context"

	tenantv1 "nexus/admin/api/build/apis/tenantruntime.admin.nexus.com/v1"

	yamlv1 "github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TenantRuntime controller", func() {
	It("should process tenant config", func() {
		common.AddTenantDisplayName("43e32e8f86ba4a5e90e44049f19ab8f73b77cc7f", "831a06ab-9781-487c-a9da-6c73973e540a")
		common.AddTenantState("831a06ab-9781-487c-a9da-6c73973e540a", common.TenantState{
			Status:        common.CREATING,
			Message:       "Tenant in provisoning",
			CreationStart: "2023-05-02T07:35:06Z",
			SKU:           "LICENSE_ADVANCE",
		})

		var obj tenantv1.Tenant
		err := yamlv1.Unmarshal([]byte(tenantRuntimeExample), &obj)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Create(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() bool {
			if tenantState, _ := common.GetTenantState("831a06ab-9781-487c-a9da-6c73973e540a"); tenantState.Message == "Apps not created" {
				return true
			}
			return false
		}, 5).Should(BeTrue())

	})
})
