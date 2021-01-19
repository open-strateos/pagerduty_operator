package pdhelpers

import (
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
)

type RulesetOptions struct {
	Name                *string
	ID                  *string
	CatchallServiceName string //required
}

type RulesetHelper struct {
	RulesetClient
}

func (rsh *RulesetHelper) AdoptOrCreateRuleset(opts *RulesetOptions) (*pagerduty.Ruleset, error) {
	// Find any existing rulesets that match the name
	resp, err := rsh.ListRulesets()
	if err != nil {
		return nil, err
	}
	matchingRulesets := make([]*pagerduty.Ruleset, 0, 4)
	for _, ruleset := range resp.Rulesets {
		if ruleset.Name == *opts.Name {
			matchingRulesets = append(matchingRulesets, ruleset)
		}
	}

	if len(matchingRulesets) > 1 {
		return nil, fmt.Errorf("%d rulesets found with name \"%s\". Don't know which one to use", len(matchingRulesets), *opts.Name)
	} else if len(matchingRulesets) == 1 {
		return matchingRulesets[0], nil
	} else {
		rs, err := rsh.createRuleset(opts)
		if err != nil {
			return nil, err
		}
		return rs, fmt.Errorf("CREATED")
	}
}

func (rsc *RulesetHelper) createRuleset(opts *RulesetOptions) (*pagerduty.Ruleset, error) {
	ruleset := &pagerduty.Ruleset{
		Name: *opts.Name,
	}
	ruleset, _, err := rsc.CreateRuleset(ruleset)
	if err != nil {
		return nil, err
	}

	rsc.addCatchallRule(ruleset, opts.CatchallServiceName)

	return ruleset, err
}

func (rsc *RulesetHelper) addCatchallRule(ruleset *pagerduty.Ruleset, targetServiceName string) error {

	// service, err := GetServiceByName(client, targetServiceName)
	// if err != nil {
	// 	return err
	// }

	// _, _, err = client.CreateRulesetRule(
	// 	ruleset.ID,
	// 	&pagerduty.RulesetRule{
	// 		CatchAll: true,
	// 		Actions: &pagerduty.RuleActions{
	// 			Route: &pagerduty.RuleActionParameter{
	// 				Value: service.ID,
	// 			},
	// 		},
	// 	},
	// )
	// return err
	return nil
}
