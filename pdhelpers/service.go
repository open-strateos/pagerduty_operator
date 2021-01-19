package pdhelpers

import (
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
)

type ServiceHelper struct {
	ServiceClient
}

func (sh *ServiceHelper) GetServiceByName(name string) (*pagerduty.Service, error) {
	resp, err := sh.ListServices(pagerduty.ListServiceOptions{
		Query: name, // underdocumented, but this appears to do a substring match on the name field
	})

	if err != nil {
		return nil, err
	}

	matches := make([]pagerduty.Service, 0, 4)
	for _, svc := range resp.Services {
		if svc.Name == name {
			matches = append(matches, svc)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("No service found with name \"%s\"", name)
	} else if len(matches) > 1 {
		return nil, fmt.Errorf("Too many services with name \"%s\" (found %d)", name, len(matches))
	} else {
		return &matches[0], nil
	}
}
