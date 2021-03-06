package controllers

import (
	"net/http"

	pd "github.com/PagerDuty/go-pagerduty"
)

const testID = "V0RB"

var okResponse = &http.Response{
	Status:     "200 OK",
	StatusCode: 200,
}

// PagerdutyClientMock mocks out the client so we can test against it.
// It stores a little bit of state so we can verify which calls were made.
type PagerdutyClientMock struct {
	service     *pd.Service
	rulesetRule *pd.RulesetRule

	updateServiceCalled bool
}

func (pdc *PagerdutyClientMock) Reset() {
	pdc.service = nil
	pdc.rulesetRule = nil
	pdc.updateServiceCalled = false
}

func (pdc *PagerdutyClientMock) GetEscalationPolicy(id string, opt *pd.GetEscalationPolicyOptions) (*pd.EscalationPolicy, error) {
	return &pd.EscalationPolicy{APIObject: pd.APIObject{ID: id}}, nil
}

func (pdc *PagerdutyClientMock) GetService(id string, opts *pd.GetServiceOptions) (*pd.Service, error) {
	return pdc.service, nil
}

func (pdc *PagerdutyClientMock) UpdateService(service pd.Service) (*pd.Service, error) {
	pdc.service = &service
	pdc.updateServiceCalled = true
	return pdc.service, nil
}

func (pdc *PagerdutyClientMock) CreateService(service pd.Service) (*pd.Service, error) {
	service.ID = testID
	pdc.service = &service
	return &service, nil
}

func (pdc *PagerdutyClientMock) GetRuleset(id string) (*pd.Ruleset, *http.Response, error) {
	return &pd.Ruleset{ID: id}, okResponse, nil
}

func (pdc *PagerdutyClientMock) GetRulesetRule(rulesetID string, ruleID string) (*pd.RulesetRule, *http.Response, error) {
	return &pd.RulesetRule{ID: ruleID}, okResponse, nil
}

func (pdc *PagerdutyClientMock) UpdateRulesetRule(rulesetID string, ruleID string, rule *pd.RulesetRule) (*pd.RulesetRule, *http.Response, error) {
	pdc.rulesetRule = rule
	return rule, okResponse, nil
}

func (pdc *PagerdutyClientMock) CreateRulesetRule(rulesetID string, rule *pd.RulesetRule) (*pd.RulesetRule, *http.Response, error) {
	rule.ID = testID
	pdc.rulesetRule = rule
	return rule, okResponse, nil
}

func (pdc *PagerdutyClientMock) DeleteRulesetRule(rulesetID string, ruleID string) error {
	pdc.rulesetRule = nil
	return nil
}

func (pdc *PagerdutyClientMock) DeleteService(id string) error {
	pdc.service = nil
	return nil
}
