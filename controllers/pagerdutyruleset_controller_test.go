package controllers

import (
	"context"
	"fmt"
	"net/http"
	v1 "pagerduty-operator/api/v1"

	"github.com/PagerDuty/go-pagerduty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PagerdutyRuleset Controller", func() {
	ctx := context.Background()
	When("Creating a new ruleset", func() {
		It("should work", func() {
			ruleset := newTestK8sRuleset("foo")
			err := k8sClient.Create(ctx, &ruleset)
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

/***
* RulesetClient Mock
***/
type MockRulesetClient struct {
	rulesetsByID map[string]*pagerduty.Ruleset
}

func NewMockRulesetClient() MockRulesetClient {
	return MockRulesetClient{
		rulesetsByID: make(map[string]*pagerduty.Ruleset),
	}
}

func (rsc MockRulesetClient) CreateRuleset(r *pagerduty.Ruleset) (*pagerduty.Ruleset, *http.Response, error) {
	if r.ID == "" {
		r.ID = RandomString(10)
	}
	rsc.rulesetsByID[r.ID] = r
	return r, &http.Response{StatusCode: http.StatusOK}, nil
}

func (rsc MockRulesetClient) DeleteRuleset(id string) error {
	var err error = nil
	if _, ok := rsc.rulesetsByID[id]; ok {
		delete(rsc.rulesetsByID, id)
	} else {
		err = fmt.Errorf("Not Found")
	}
	return err
}

func (rsc MockRulesetClient) GetRuleset(id string) (*pagerduty.Ruleset, *http.Response, error) {
	rs, ok := rsc.rulesetsByID[id]
	var statusCode int
	var err error = nil
	if ok {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusNotFound
		err = fmt.Errorf("Not Found")
	}
	return rs, &http.Response{StatusCode: statusCode}, err
}

func (rsc MockRulesetClient) ListRulesets() (*pagerduty.ListRulesetsResponse, error) {
	size := uint(len(rsc.rulesetsByID))
	rulesets := make([]*pagerduty.Ruleset, 0, size)
	for _, v := range rsc.rulesetsByID {
		rulesets = append(rulesets, v)
	}
	resp := pagerduty.ListRulesetsResponse{
		Total:    size,
		Rulesets: rulesets,
	}
	return &resp, nil
}

func (rsc MockRulesetClient) UpdateRuleset(r *pagerduty.Ruleset) (*pagerduty.Ruleset, *http.Response, error) {
	rsc.rulesetsByID[r.ID] = r
	return r, &http.Response{StatusCode: http.StatusOK}, nil
}
