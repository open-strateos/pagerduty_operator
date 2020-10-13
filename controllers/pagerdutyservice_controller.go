/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/dchest/uniuri"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "pagerduty-operator/api/v1"
	v1 "pagerduty-operator/api/v1"
)

const finalizerKey = "pagerdutyservice.core.strateos.com"

// PagerdutyServiceReconciler reconciles a PagerdutyService object
type PagerdutyServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	PdClient      PagerdutyInterface
	RulesetID     string
	ServicePrefix string // append to service names
}

var logger = ctrl.Log.WithName("pagerdutyServiceReconciler")

// +kubebuilder:rbac:groups=core.strateos.com,resources=pagerdutyservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.strateos.com,resources=pagerdutyservices/status,verbs=get;update;patch

func (r *PagerdutyServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("pagerdutyservice", req.NamespacedName)

	var kubeService v1.PagerdutyService
	var pdService *pagerduty.Service
	var err error

	logger.V(1).Info("Fetching PagerdutyService resource")
	if err = r.Get(ctx, req.NamespacedName, &kubeService); err != nil {
		logger.V(1).Info("Unable to fetch PagerdutyService")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	spec := &kubeService.Spec
	status := &kubeService.Status

	if kubeService.DeletionTimestamp.IsZero() {
		kubeService.EnsureFinalizerExists(finalizerKey)
	} else {
		logger.Info("Resource is marked for deletion. Cleaning up.")
		err = r.destroyPagerdutyResources(&kubeService)
		if err == nil {
			// when everything is cleaned up, remove the finalizer, so k8s can delete the resource
			logger.Info("Cleanup succesful")
			kubeService.EnsureFinalizerRemoved(finalizerKey)
			err = r.Update(ctx, kubeService.DeepCopyObject())
		}
		return ctrl.Result{}, err
	}

	var escalationPolicy *pagerduty.EscalationPolicy

	escalationPolicy, err = r.PdClient.GetEscalationPolicy(spec.EscalationPolicy, &pagerduty.GetEscalationPolicyOptions{})
	if escalationPolicy == nil {
		delay := time.Second * 30
		logger.Error(err, "Can't find the escalation policy. Will retry.", "policyID", spec.EscalationPolicy, "delay", delay)
		return ctrl.Result{Requeue: true, RequeueAfter: delay}, nil
	}

	var serviceExists bool
	if status.ServiceID != "" { // Service might already exist
		logger.Info("Fetching service from pagerduty", "serviceId", status.ServiceID, "serviceName", status.ServiceName)
		pdService, err = r.PdClient.GetService(status.ServiceID, &pagerduty.GetServiceOptions{})
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		serviceExists = pdService != nil
	}
	if !serviceExists {
		pdService = &pagerduty.Service{}
	}

	pdService.Description = spec.Description
	pdService.EscalationPolicy = *escalationPolicy

	if serviceExists {
		pdService, err = r.PdClient.UpdateService(*pdService)
	} else {
		pdService.Name = r.generatePdServiceName(kubeService.Name, 0)
		pdService, err = r.PdClient.CreateService(*pdService)
	}
	if err != nil {
		logger.Error(err, "Failed to create pagerduty service resource", "service", pdService)
		return ctrl.Result{}, err
	}
	kubeService.Status.ServiceID = pdService.ID
	kubeService.Status.ServiceName = pdService.Name

	r.reconcileRoutingRules(&kubeService)

	err = r.Update(ctx, kubeService.DeepCopyObject())
	return ctrl.Result{}, err
}

func (r *PagerdutyServiceReconciler) reconcileRoutingRules(kubeService *corev1.PagerdutyService) error {
	ruleset, _, err := r.PdClient.GetRuleset(r.RulesetID)
	if err != nil {
		return err
	}

	var rule *pagerduty.RulesetRule
	ruleID := kubeService.Status.RuleID
	ruleExists := ruleID != ""

	if !ruleExists {
		logger.Info("Creating new rule")
		rule = &pagerduty.RulesetRule{
			Ruleset: &pagerduty.APIObject{
				ID: ruleset.ID,
			},
		}
	} else {
		logger.V(1).Info("Using existing rule")
		rule, _, err = r.PdClient.GetRulesetRule(ruleset.ID, ruleID)
		if err != nil {
			return err
		}
	}

	conditions := pagerduty.RuleConditions{
		Operator: "and",
	}
	for _, labelSpec := range kubeService.Spec.MatchLabels {
		subcondition := pagerduty.RuleSubcondition{
			Operator: "contains",
			Parameters: &pagerduty.ConditionParameter{
				Path:  "details.firing",
				Value: fmt.Sprintf("%s = %s", labelSpec.Key, labelSpec.Value),
			},
		}
		conditions.RuleSubconditions = append(conditions.RuleSubconditions, &subcondition)
	}
	rule.Conditions = &conditions

	serviceID := kubeService.Status.ServiceID
	rule.Actions = &pagerduty.RuleActions{
		Route: &pagerduty.RuleActionParameter{Value: serviceID},
	}

	if ruleExists {
		rule, _, err = r.PdClient.UpdateRulesetRule(ruleset.ID, rule.ID, rule)
		logger.Info("Updated routing rule", "rule", rule)
	} else {
		rule, _, err = r.PdClient.CreateRulesetRule(ruleset.ID, rule)
		logger.Info("Created routing rule", "rule", rule)
	}

	if err != nil {
		return err
	}

	kubeService.Status.RuleID = rule.ID

	return nil
}

func (r *PagerdutyServiceReconciler) destroyPagerdutyResources(kubeService *corev1.PagerdutyService) error {
	logger.Info("Resource is marked for deletion. Cleaning up.")
	var err error

	ruleID := kubeService.Status.RuleID
	if ruleID != "" {
		err := r.PdClient.DeleteRulesetRule(r.RulesetID, ruleID)
		if err != nil {
			return err
		}
		logger.Info("Successfully deleted the routing rule")
	}

	serviceID := kubeService.Status.ServiceID
	if serviceID != "" {
		err = r.PdClient.DeleteService(kubeService.Status.ServiceID)
		if err != nil {
			return err
		}
		logger.Info("Successfully deleted the pagerduty service")
	}

	return nil
}

// generatePdServiceName prepends the configured prefix if applicable
// it will also add a random suffix of a given length (to overcome pagerduty's flat namespace for service names)
func (r *PagerdutyServiceReconciler) generatePdServiceName(name string, randomSuffixLen int) string {
	if r.ServicePrefix != "" {
		name = r.ServicePrefix + "-" + name
	}
	if randomSuffixLen > 0 {
		name = name + "-" + uniuri.NewLen(randomSuffixLen)
	}
	return name
}

func (r *PagerdutyServiceReconciler) escalationPolicyExists(policyId string) (bool, error) {
	policy, err := r.PdClient.GetEscalationPolicy(policyId, &pagerduty.GetEscalationPolicyOptions{})
	if policy != nil {
		return true, nil
	}
	return false, err
}

func (r *PagerdutyServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PagerdutyService{}).
		Complete(r)
}

// PagerdutyInterface allows us to write a fake client for testing
type PagerdutyInterface interface {
	GetEscalationPolicy(id string, opt *pagerduty.GetEscalationPolicyOptions) (*pagerduty.EscalationPolicy, error)
	GetService(id string, opts *pagerduty.GetServiceOptions) (*pagerduty.Service, error)
	UpdateService(service pagerduty.Service) (*pagerduty.Service, error)
	CreateService(service pagerduty.Service) (*pagerduty.Service, error)
	GetRuleset(id string) (*pagerduty.Ruleset, *http.Response, error)
	GetRulesetRule(ruleID string, rulesetID string) (*pagerduty.RulesetRule, *http.Response, error)
	UpdateRulesetRule(ruleID string, rulesetID string, rule *pagerduty.RulesetRule) (*pagerduty.RulesetRule, *http.Response, error)
	CreateRulesetRule(ruleID string, rule *pagerduty.RulesetRule) (*pagerduty.RulesetRule, *http.Response, error)
	DeleteRulesetRule(ruleID string, rulesetID string) error
	DeleteService(id string) error
}
