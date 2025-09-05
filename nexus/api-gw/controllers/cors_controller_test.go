package controllers

import (
	"api-gw/pkg/model"
	"context"
	"encoding/json"

	domain_nexus_org "nexus/admin/api/build/apis/domain.admin.nexus.com/v1"

	yamlv1 "github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OidcConfig controller", func() {
	It("should process oidc config", func() {
		crdJson, err := yamlv1.YAMLToJSON([]byte(corsConfigExample))
		Expect(err).NotTo(HaveOccurred())

		var obj domain_nexus_org.CORSConfig
		err = json.Unmarshal(crdJson, &obj)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Create(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())

		event := <-model.CorsChan
		Expect(event.Type).To(Equal(model.Upsert))
		Expect(event.Cors.Name).To(Equal("default"))

		err = k8sClient.Delete(context.TODO(), &obj)
		Expect(err).ToNot(HaveOccurred())

		event = <-model.CorsChan
		Expect(event.Type).To(Equal(model.Delete))
	})
})
