package pdhelpers

import (
	"fmt"
	"net/http"

	"github.com/PagerDuty/go-pagerduty"
)

type RulesetHelper struct {
	RulesetClient
}

// AdoptOrCreateRuleset either fetches or create a ruleset matching the supplied options
// Returns a pointer to the ruleset, a boolean indicating whether a new resource was created,
// and an optional error
func (rsh *RulesetHelper) AdoptOrCreateRuleset(name string) (*pagerduty.Ruleset, bool, error) {
	// Find any existing rulesets that match the name
	existingRuleset, err := rsh.GetRulesetByName(name)
	if err != nil {
		return nil, false, err
	}

	if existingRuleset == nil {
		// No ruleset found. Create one.
		rs, resp, err := rsh.CreateRuleset(&pagerduty.Ruleset{
			Name: name,
		})
		if resp.StatusCode != 201 {
			return nil, false, fmt.Errorf("Error code %d while creating ruleset: %s", resp.StatusCode, resp.Status)
		} else if rs == nil {
			return nil, false, fmt.Errorf("ruleset was nil, but no errors from API")
		} else if err != nil {
			return nil, false, err
		}
		return rs, true, err
	} else {
		// adopt existing ruleset
		return existingRuleset, false, nil
	}

}

func (rsh *RulesetHelper) GetRulesetByName(name string) (*pagerduty.Ruleset, error) {
	resp, err := rsh.ListRulesets()
	if err != nil {
		return nil, err
	}
	matchingRulesets := make([]*pagerduty.Ruleset, 0, 1)
	for _, ruleset := range resp.Rulesets {
		if ruleset.Name == name {
			matchingRulesets = append(matchingRulesets, ruleset)
		}
	}

	if len(matchingRulesets) < 1 {
		return nil, nil // no error, but no ruleset exists
	} else if len(matchingRulesets) > 1 {
		return nil, fmt.Errorf("Didn't expect to find %d rulesets named %v", len(matchingRulesets), name)
	}
	return matchingRulesets[0], nil

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

/***
* FakeRulesetClient, for testing
***/
type FakeRulesetClient struct {
	RulesetsByID map[string]*pagerduty.Ruleset
}

func NewFakeRulesetClient() FakeRulesetClient {
	rsc := FakeRulesetClient{RulesetsByID: make(map[string]*pagerduty.Ruleset)}
	return rsc
}

func (rsc FakeRulesetClient) Reset() {
	// clear the "database"
	rsc.RulesetsByID = make(map[string]*pagerduty.Ruleset)
}

func (rsc FakeRulesetClient) CreateRuleset(r *pagerduty.Ruleset) (*pagerduty.Ruleset, *http.Response, error) {
	if r.ID == "" {
		r.ID = RandomString(10)
	}
	rsc.RulesetsByID[r.ID] = r
	return r, &http.Response{StatusCode: http.StatusCreated}, nil
}

func (rsc FakeRulesetClient) DeleteRuleset(id string) error {
	var err error = nil
	if _, ok := rsc.RulesetsByID[id]; ok {
		delete(rsc.RulesetsByID, id)
	} else {
		err = fmt.Errorf("Not Found")
	}
	return err
}

func (rsc FakeRulesetClient) GetRuleset(id string) (*pagerduty.Ruleset, *http.Response, error) {
	rs, ok := rsc.RulesetsByID[id]
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

func (rsc FakeRulesetClient) ListRulesets() (*pagerduty.ListRulesetsResponse, error) {
	size := uint(len(rsc.RulesetsByID))
	rulesets := make([]*pagerduty.Ruleset, 0, size)
	for _, v := range rsc.RulesetsByID {
		rulesets = append(rulesets, v)
	}
	resp := pagerduty.ListRulesetsResponse{
		Total:    size,
		Rulesets: rulesets,
	}
	return &resp, nil
}

func (rsc FakeRulesetClient) UpdateRuleset(r *pagerduty.Ruleset) (*pagerduty.Ruleset, *http.Response, error) {
	rsc.RulesetsByID[r.ID] = r
	return r, &http.Response{StatusCode: http.StatusOK}, nil
}
