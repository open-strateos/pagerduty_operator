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
	When("Creating a new ruleset", func() {
		rulesetName := "foo"
		It("should work", func() {
			ruleset := newTestK8sRuleset(rulesetName)
			err := k8sClient.Create(ctx, &ruleset)
			Expect(err).ToNot(HaveOccurred())

			rsh := pdhelpers.RulesetHelper{RulesetClient: fakeRulesetClient}

			Eventually(func() bool {
				rulesets, err := rsh.GetRulesetsByName(rulesetName)
				if err != nil || len(rulesets) < 1 {
					return false
				}
				return rulesets[0].Name == rulesetName
			}).Should(BeTrue())
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
