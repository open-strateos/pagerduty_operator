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

// AdoptOrCreateRuleset either fetches or create a ruleset matching the supplied options
// Returns a pointer to the ruleset, a boolean indicating whether a new resource was created,
// and an optional error
func (rsh *RulesetHelper) AdoptOrCreateRuleset(opts *RulesetOptions) (*pagerduty.Ruleset, bool, error) {
	// Find any existing rulesets that match the name
	resp, err := rsh.ListRulesets()
	if err != nil {
		return nil, false, err
	}
	matchingRulesets := make([]*pagerduty.Ruleset, 0, 4)
	for _, ruleset := range resp.Rulesets {
		if ruleset.Name == *opts.Name {
			matchingRulesets = append(matchingRulesets, ruleset)
		}
	}

	if len(matchingRulesets) > 1 {
		return nil, false, fmt.Errorf("%d rulesets found with name \"%s\". Don't know which one to use", len(matchingRulesets), *opts.Name)
	} else if len(matchingRulesets) == 1 {
		return matchingRulesets[0], false, nil
	} else {
		rs, err := rsh.createRuleset(opts)
		return rs, true, err
	}
}

func (rsh *RulesetHelper) createRuleset(opts *RulesetOptions) (*pagerduty.Ruleset, error) {
	ruleset := &pagerduty.Ruleset{
		Name: *opts.Name,
	}
	ruleset, _, err := rsh.CreateRuleset(ruleset)
	if err != nil {
		return nil, err
	}

	rsh.addCatchallRule(ruleset, opts.CatchallServiceName)

	return ruleset, err
}

func (rsh *RulesetHelper) addCatchallRule(ruleset *pagerduty.Ruleset, targetServiceName string) error {

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
