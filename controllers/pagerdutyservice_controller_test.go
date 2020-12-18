package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pagerdutyAPIV1 "pagerduty-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	timeout := "3s"
	interval := "10ms"

	ctx := context.Background()
	var pdService *pagerdutyAPIV1.PagerdutyService
	var serviceNamespacedName types.NamespacedName
	var err error

	When("Creating a PagerdutyService", func() {

		pdClientMock.Reset()
		It("Should start with nil values for service and rule", func() {
			Expect(pdClientMock.service).Should(BeNil())
			Expect(pdClientMock.rulesetRule).Should(BeNil())

			//but not k8sClient
			Expect(k8sClient).ToNot(BeNil())
		})

		It("Should create successfully", func() {
			pdService, err = loadPdServiceFromYaml(pdServiceYaml)
			serviceNamespacedName = getNamespacedName(pdService)
			Expect(err).ShouldNot(HaveOccurred())
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

	When("Updating the PagerdutyService", func() {

		It("Should successfully update", func() {
			// Sometimes the "fetch, modify, update" cycle requires retry to avoid
			// a ResourceVersion conflict, presumably because the reconciler is updating
			// the Status fields in the meantime.
			Eventually(func() error {
				updatedPdService := pagerdutyAPIV1.PagerdutyService{}
				Expect(k8sClient.Get(ctx, serviceNamespacedName, &updatedPdService)).To(Succeed())
				updatedPdService.Spec.Description = "aaaaa"
				updatedPdService.Spec.EscalationPolicy = "bbbbb"
				return k8sClient.Update(ctx, &updatedPdService)
			}, timeout, interval).Should(Succeed())
		})

		It("Should also update the service in pagerduty", func() {

			Eventually(func() bool {
				return pdClientMock.updateServiceCalled
			}).Should(BeTrue())

			Eventually(func() string {
				return pdClientMock.service.Description
			}).Should(Equal(pdService.Spec.Description))

			Eventually(func() string {
				return pdClientMock.service.EscalationPolicy.ID
			}).Should(Equal(pdService.Spec.EscalationPolicy))
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

	When("Lookup excalation policy from a Secret", func() {
		ctx := context.Background()

		const secretName = "some-secret"
		const secretKey = "some-key"
		const escalationPolicyID = "1234ABC"

		// Create the test secret
		testSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: "default",
			},
			StringData: map[string]string{
				secretKey: escalationPolicyID,
			},
		}

		It("Should create the secret no problem", func() {
			Expect(k8sClient.Create(ctx, &testSecret)).Should(Succeed())
		})

		serviceYamlWithSecretNameAndKey := fmt.Sprintf(`---
apiVersion: "core.strateos.com/v1"
kind: PagerdutyService
metadata:
  name: test-service-bravo
  namespace: default
spec:	
    description: Testing the operator
    escalationPolicySecret:
        name: %s
        key: %s
    matchLabels:
        - key: foo
          value: bar
    `, secretName, secretKey)

		When("Creating a pagerduty service with escalation policy from a Secret name and key", func() {
			pdService, err := loadPdServiceFromYaml(serviceYamlWithSecretNameAndKey)

			It("Should be able to fetch the secret", func() {
				Expect(err).ToNot(HaveOccurred())
				newSecret := corev1.Secret{}
				objectKey, _ := runtimeClient.ObjectKeyFromObject(&testSecret)
				k8sClient.Get(ctx, objectKey, &newSecret)
				Expect(newSecret.Name).To(Equal(testSecret.Name))
			})

			It("Reconciler should be able to extract escalation policy from the secret", func() {
				id, err := pagerdutyServiceReconciler.GetEscalationPolicyID(pdService)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(id).Should(Equal(escalationPolicyID))
			})

			It("Should be able to create a PagerdutyService resource without error.", func() {
				Expect(k8sClient.Create(ctx, pdService)).Should(Succeed())
			})

		})

	})

})

func getNamespacedName(service *pagerdutyAPIV1.PagerdutyService) types.NamespacedName {
	key, err := runtimeClient.ObjectKeyFromObject(service)
	Expect(err).NotTo(HaveOccurred())
	return key
}

func loadPdServiceFromYaml(yamlRep string) (*pagerdutyAPIV1.PagerdutyService, error) {
	decoder := scheme.Codecs.UniversalDeserializer()
	pdService := &pagerdutyAPIV1.PagerdutyService{}
	_, _, err := decoder.Decode([]byte(yamlRep), nil, pdService)

	return pdService, err
}
