package controllers

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pagerdutyAPIV1 "pagerduty-operator/api/v1"
)

const (
	pagerdutyServiceKind = "PagerdutyService"
	rulesetID            = "WJGIH"
	servicePrefix        = "whatever"
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
	// Timing parameters for "Eventually" polling
	timeout := "1s"
	interval := "10ms"

	ctx := context.Background()
	var pdService *pagerdutyAPIV1.PagerdutyService

	When("Creating a PagerdutyService", func() {

		var serviceNamespacedName types.NamespacedName

		It("Should start with nil values for service and rule", func() {
			Expect(pdClientMock.service).Should(BeNil())
			Expect(pdClientMock.rulesetRule).Should(BeNil())
		})

		It("Should create successfully", func() {
			pdService = loadPdServiceFromYaml(pdServiceYaml)
			serviceNamespacedName = getNamespacedName(pdService)
			Expect(pdService.Name).ShouldNot(BeEmpty())
			Expect(pdService.Namespace).Should(BeIdenticalTo("default"))
			Expect(k8sClient.Create(ctx, pdService)).Should(Succeed())
		})

		It("Should eventually create a service", func() {
			Eventually(func() bool {
				return pdClientMock.service != nil
			}, timeout, interval).Should(BeTrue())
		})

		It("Should have a correctly prefixed name", func() {
			Eventually(func() string {
				return pdClientMock.service.Name
			}, timeout, interval).Should(Equal(fmt.Sprintf("%s-%s", servicePrefix, pdService.Name)))
		})

		It("Should eventually create a rule", func() {
			Eventually(func() bool {
				return pdClientMock.rulesetRule != nil
			}, timeout, interval).Should(BeTrue())
		})

		It("Should update the PagerdutyService resource status", func() {
			Eventually(func() string {
				service := &pagerdutyAPIV1.PagerdutyService{}
				Expect(k8sClient.Get(ctx, serviceNamespacedName, service)).To(Succeed())
				return service.Status.ServiceID
			}, timeout, interval).Should(Equal(testID))

			Eventually(func() string {
				service := &pagerdutyAPIV1.PagerdutyService{}
				Expect(k8sClient.Get(ctx, serviceNamespacedName, service)).To(Succeed())
				return service.Status.RuleID
			}, timeout, interval).Should(Equal(testID))
		})

		It("Should have a finalizer", func() {
			Eventually(func() int {
				service := &pagerdutyAPIV1.PagerdutyService{}
				Expect(k8sClient.Get(ctx, serviceNamespacedName, service)).To(Succeed())
				return len(service.ObjectMeta.Finalizers)
			}, timeout, interval).Should(BeNumerically(">", 0))
		})
	})

	When("Deleting PagerdutyService", func() {
		It("Mock should intially have populated resources", func() {
			Expect(pdClientMock.service).ToNot(BeNil())
			Expect(pdClientMock.rulesetRule).ToNot(BeNil())
		})

		It("Should deleteSuccesfully", func() {
			Expect(k8sClient.Delete(ctx, pdService)).To(Succeed())
		})

		It("Should eventually delete the PD service", func() {
			Eventually(func() bool {
				return pdClientMock.service == nil
			}, timeout, interval).Should(BeTrue())
		})

		It("Should eventually delete the PD Rule", func() {
			Eventually(func() bool {
				return pdClientMock.rulesetRule == nil
			}, timeout, interval).Should(BeTrue())
		})

		It("Should eventually delete the resource", func() {
			Eventually(func() error {
				service := &pagerdutyAPIV1.PagerdutyService{}
				return k8sClient.Get(ctx, getNamespacedName(pdService), service)
			}, timeout, interval).Should(HaveOccurred())
		})
	})
})

func getNamespacedName(service *pagerdutyAPIV1.PagerdutyService) types.NamespacedName {
	key, err := runtimeClient.ObjectKeyFromObject(service)
	if err != nil {
		log.Fatal(err)
	}
	return key
}

func loadPdServiceFromYaml(yamlRep string) *pagerdutyAPIV1.PagerdutyService {
	var decoder = scheme.Codecs.UniversalDeserializer()
	obj, gkv, err := decoder.Decode([]byte(yamlRep), nil, nil)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(gkv.Kind).Should(Equal(pagerdutyServiceKind))
	pdService := obj.(*pagerdutyAPIV1.PagerdutyService)
	return pdService
}
