package controllers

import (
	"context"
	v1 "pagerduty-operator/api/v1"
	"pagerduty-operator/pdhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PagerdutyRuleset Controller", func() {
	ctx := context.Background()
	rulesetName := "foo"
	testRuleset := newTestK8sRuleset(rulesetName)

	When("Creating a new ruleset", func() {
		It("should work", func() {
			err := k8sClient.Create(ctx, &testRuleset)
			Expect(err).ToNot(HaveOccurred())

			rsh := pdhelpers.RulesetHelper{RulesetClient: fakeRulesetClient}

			Eventually(func() bool {
				ruleset, err := rsh.GetRulesetByName(rulesetName)
				if ruleset == nil || err != nil {
					return false
				}
				return ruleset.Name == rulesetName
			}).Should(BeTrue())
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
})

func newTestK8sRuleset(name string) v1.PagerdutyRuleset {
	return v1.PagerdutyRuleset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
	}
}
