package controllers

import (
	"context"
	v1 "pagerduty-operator/api/v1"
	"pagerduty-operator/pdhelpers"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("PagerdutyRuleset Controller", func() {
	ctx := context.Background()
	rulesetName := "foo"
	testRuleset := newTestK8sRuleset(rulesetName)
	testRulesetNamespacedName := types.NamespacedName{
		Namespace: testRuleset.GetObjectMeta().GetNamespace(),
		Name:      testRuleset.GetObjectMeta().GetName(),
	}

	When("Creating a new ruleset", func() {
		It("should work", func() {
			err := k8sClient.Create(ctx, &testRuleset)
			Expect(err).ToNot(HaveOccurred())

			rsh := pdhelpers.RulesetHelper{RulesetClient: fakeRulesetClient}

			var ruleset *pagerduty.Ruleset
			Eventually(func() bool {
				ruleset, err = rsh.GetRulesetByName(rulesetName)
				if ruleset == nil || err != nil {
					return false
				}
				return ruleset.Name == rulesetName
			}).Should(BeTrue())

			var createdRuleset v1.PagerdutyRuleset
			err = k8sClient.Get(ctx, testRulesetNamespacedName, &createdRuleset)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() string {
				_ = k8sClient.Get(ctx, testRulesetNamespacedName, &createdRuleset)
				return createdRuleset.Status.RulesetID
			}).Should(Equal(ruleset.ID))
			Expect(createdRuleset.Status.Created).To(BeTrue())
		})
	})

	When("Deleting a ruleset", func() {
		It("Should clean up the pagerduty rulesest", func() {
			testRulesetID := testRuleset.Status.RulesetID
			Expect(len(fakeRulesetClient.RulesetsByID)).To(Equal(1))
			// Expect(fakeRulesetClient.RulesetsByID[testRulesetID].Name).To(Equal(testRuleset.Name))

			err := k8sClient.Delete(ctx, &testRuleset)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				_, exists := fakeRulesetClient.RulesetsByID[testRulesetID]
				return exists
			}).Should(BeFalse())

		})
	})

	When("checking package-global variables", func() {
		It("srsly", func() {
			Expect(k8sClient).NotTo(BeNil())
			Expect(fakeRulesetClient.RulesetsByID).NotTo(BeNil())
		})
	})

	When("Adopting an existing ruleset", func() {
		adoptedRulesetName := "already-exists"
		It("Does the right thing.", func() {

			// Make sure Ginkgo isn't doing shady things with execution order
			Expect(fakeRulesetClient.RulesetsByID).ToNot(BeNil())
			Expect(k8sClient).NotTo(BeNil())

			// A pre-existing ruleset in the pagerduty API
			existingRuleset := &pagerduty.Ruleset{Name: adoptedRulesetName}
			existingRuleset, _, err := fakeRulesetClient.CreateRuleset(existingRuleset)
			Expect(err).NotTo(HaveOccurred())

			// Create a k8s ruleset with the same name
			adoptedRuleset := newTestK8sRuleset(adoptedRulesetName)
			err = k8sClient.Create(ctx, &adoptedRuleset)
			Expect(err).ToNot(HaveOccurred())
			namespacedName := types.NamespacedName{
				Namespace: adoptedRuleset.Namespace,
				Name:      adoptedRuleset.Name,
			}

			// Wait for reconcile
			Eventually(func() error {
				return k8sClient.Get(ctx, namespacedName, &adoptedRuleset)
			}).ShouldNot(HaveOccurred())
			Eventually(func() string {
				_ = k8sClient.Get(ctx, namespacedName, &adoptedRuleset)
				return adoptedRuleset.Status.RulesetID
			}).ShouldNot(BeEmpty())

			// Shold be marked as adopted
			rulesetID := adoptedRuleset.Status.RulesetID
			Expect(adoptedRuleset.Status.Created).To(BeFalse())
			Expect(rulesetID).ToNot(BeEmpty())

			// Delete the k8s ruleset
			err = k8sClient.Delete(ctx, &adoptedRuleset)
			Expect(err).ToNot(HaveOccurred())

			// Wait for cleanup to finish
			Eventually(func() error {
				return k8sClient.Get(ctx, namespacedName, &adoptedRuleset)
			}, "3s").Should(HaveOccurred())

			// pagerdyty ruleset should still exist
			r, _, err := fakeRulesetClient.GetRuleset(rulesetID)
			Expect(r).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

	})
})

func newTestK8sRuleset(name string) v1.PagerdutyRuleset {
	return v1.PagerdutyRuleset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
	}
}
