pagerduty-operator
==================

pagerduty-operator, and its associated `PagerdutyService` CRD
can manage Pagerduty services and routing rules, in a fairly
simplistic way, by bridging declarative Kubernetes resources
with the Pagerduty REST API.

An instance of pagerduty-operator is configured to manage a single
pager duty global ruleset (specified by the `PAGERDUTY_RULESET_ID` environment variable or the `-ruleset` runtime flag). For each `PagerdutyService` resource the operator sees, it will create a corresponding service via the pagerduty API, and a ruleset rule that routes to that service based on alert labels.

Operator Runtime Flags
----------------------
```
  -api-key string (Default: $PAGERDUTY_API_KEY)
    	Authorization key for the pagerduty API.
  -enable-leader-election
    	Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
  -kubeconfig string
    	Paths to a kubeconfig. Only required if out-of-cluster.
  -metrics-addr string (Default: $METRICS_ADDR or ":8080")
    	The address the metric endpoint binds to.
  -ruleset string (Default: $PAGERDUTY_RULESET_ID)
    	ID of the ruleset to append routing rules to.
  -service-prefix string (Default: $PAGERDUTY_SERVICE_PREFIX)
    	Prefix to be added to Pagerduty Service names
```

Example
-------

```yaml
---
apiVersion: core.strateos.com/v1
kind: PagerdutyService
metadata:
  name: turboencabulator
spec:
  description: Instead of power being generated by the relative motion of conductors and fluxes, it is produced by the modial interaction of magneto-reluctance and capacitive diractance
  escalationPolicy: PDAVWNR # ID of an escalation policy that must already exist in pagerduty
  matchLabels:
      - key: pdService
        value: turboencabulator
```

If the operator is running with `-service-prefix foo`,
the above resource manifest will cause it to create
a service named `foo-turboencabulator` and route to it 
any incoming alerts with the label `pdService: turboencabulator`.

If the manifest is deleted, the operator will clean up both the service
and the routing rule.