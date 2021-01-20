package pdhelpers

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestRulesetHelpers(t *testing.T) {
	g := NewGomegaWithT(t)
	fakeClient := NewFakeRulesetClient()
	g.Expect(fakeClient.RulesetsByID).NotTo(BeNil())
	rsHelper := RulesetHelper{RulesetClient: fakeClient}
	rulesetName := "foo"

	// Should create as expected
	rs, created, err := rsHelper.AdoptOrCreateRuleset(rulesetName)
	g.Expect(rs).NotTo(BeNil())
	g.Expect(created).To(BeTrue())
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(rs.Name).To(Equal(rulesetName))
	g.Expect(rs.ID).NotTo(BeNil())

	// Second call should yield the same ruleset
	rs2, created, err := rsHelper.AdoptOrCreateRuleset(rulesetName)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(created).To(BeFalse())
	g.Expect(rs2.ID).To(Equal(rs.ID))

	g.Expect(len(fakeClient.RulesetsByID)).To(Equal(1))

}
