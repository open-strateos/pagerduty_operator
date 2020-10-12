package controllers

import (
	"context"

	"k8s.io/client-go/kubernetes/scheme"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pagerdutyAPIV1 "pagerduty-operator/api/v1"
)

const (
	pagerdutyServiceKind = "PagerdutyService"
)

const pdServiceYaml = `---
apiVersion: core.strateos.com/v1
kind: PagerdutyService
metadata:
  name: test-service
  namespace: default
spec:	
  description: Testing the operator
  escalationPolicy: PDAVWNR
  matchLabels:
      - key: foo
        value: bar
      - key: fnord
        value: whatever
`

var _ = Describe("PagerdutyService controller", func() {
	Context("When creating a PagerdutyService", func() {
		It("Should create successfully", func() {
			ctx := context.Background()

			pdService := loadPdServiceFromYaml(pdServiceYaml)
			Expect(pdService.Name).ShouldNot(BeEmpty())
			Expect(pdService.Namespace).Should(BeIdenticalTo("default"))
			Expect(k8sClient.Create(ctx, pdService)).Should(Succeed())
		})
	})
})

func loadPdServiceFromYaml(yamlRep string) *pagerdutyAPIV1.PagerdutyService {
	var decoder = scheme.Codecs.UniversalDeserializer()
	obj, gkv, err := decoder.Decode([]byte(pdServiceYaml), nil, nil)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(gkv.Kind).Should(Equal(pagerdutyServiceKind))
	pdService := obj.(*pagerdutyAPIV1.PagerdutyService)
	return pdService
}
