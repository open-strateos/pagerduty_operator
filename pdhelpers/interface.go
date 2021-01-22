package pdhelpers

import (
	"net/http"

	"github.com/PagerDuty/go-pagerduty"
)

// PagerdutyInterface allows us to write a fake client for testing
type PagerdutyClientInterface interface {
	EscalationPolicyClient
	RulesetClient
	RulesetRuleClient
	ServiceClient
}

var pdci PagerdutyClientInterface = (*pagerduty.Client)(nil)

// var _ RulesetClient = (*PagerdutyClientInterface)(nil)

type ServiceClient interface {
	CreateService(service pagerduty.Service) (*pagerduty.Service, error)
	DeleteService(id string) error
	GetService(id string, opts *pagerduty.GetServiceOptions) (*pagerduty.Service, error)
	ListServices(o pagerduty.ListServiceOptions) (*pagerduty.ListServiceResponse, error)
	UpdateService(service pagerduty.Service) (*pagerduty.Service, error)
}

var _ ServiceClient = (*pagerduty.Client)(nil) // ensure pagerduty client meets the interface

// RulesetClient is just the components of pagerduty.Client that involve rulesets
type RulesetClient interface {
	CreateRuleset(r *pagerduty.Ruleset) (*pagerduty.Ruleset, *http.Response, error)
	DeleteRuleset(id string) error
	GetRuleset(id string) (*pagerduty.Ruleset, *http.Response, error)
	ListRulesets() (*pagerduty.ListRulesetsResponse, error)
	UpdateRuleset(r *pagerduty.Ruleset) (*pagerduty.Ruleset, *http.Response, error)
}

var _ RulesetClient = (*pagerduty.Client)(nil) // Ensure published client matches this interface.

type RulesetRuleClient interface {
	CreateRulesetRule(ruleID string, rule *pagerduty.RulesetRule) (*pagerduty.RulesetRule, *http.Response, error)
	DeleteRulesetRule(ruleID string, rulesetID string) error
	GetRulesetRule(ruleID string, rulesetID string) (*pagerduty.RulesetRule, *http.Response, error)
	ListRulesetRules(rulesetID string) (*pagerduty.ListRulesetRulesResponse, error)
	UpdateRulesetRule(ruleID string, rulesetID string, rule *pagerduty.RulesetRule) (*pagerduty.RulesetRule, *http.Response, error)
}

var _ RulesetRuleClient = (*pagerduty.Client)(nil)

type EscalationPolicyClient interface {
	GetEscalationPolicy(id string, opt *pagerduty.GetEscalationPolicyOptions) (*pagerduty.EscalationPolicy, error)
}

var _ EscalationPolicyClient = (*pagerduty.Client)(nil)
