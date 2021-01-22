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
	"strings"
	"time"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/dchest/uniuri"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "pagerduty-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
)

const finalizerKey = "pagerdutyservice.core.strateos.com"

// PagerdutyServiceReconciler reconciles a PagerdutyService object
type PagerdutyServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	PdClient      ServiceReconcilerPagerdutyInterface
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
		EnsureFinalizerExists(&kubeService.ObjectMeta, finalizerKey)
	} else {
		logger.Info("Resource is marked for deletion. Cleaning up.")
		err = r.destroyPagerdutyResources(&kubeService)
		if err == nil {
			// when everything is cleaned up, remove the finalizer, so k8s can delete the resource
			logger.Info("Cleanup succesful")
			EnsureFinalizerRemoved(&kubeService.ObjectMeta, finalizerKey)
			err = r.Update(ctx, kubeService.DeepCopyObject())
		}
		return ctrl.Result{}, err
	}

	escalationPolicyID, err := r.GetEscalationPolicyID(&kubeService)
	if err != nil {
		logger.Info("Could not resolve the escalation policy ID. Will retry.", "pdService", kubeService.Name)
		r.UpdateStatus(&kubeService, err)
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 30}, nil
	}

	escalationPolicy, err := r.PdClient.GetEscalationPolicy(escalationPolicyID, &pagerduty.GetEscalationPolicyOptions{})
	if escalationPolicy == nil {
		delay := time.Second * 30
		logger.Error(err, "Can't find the escalation policy. Will retry.", "policyID", spec.EscalationPolicy, "delay", delay)
		r.UpdateStatus(&kubeService, fmt.Errorf("Unable to get the escaltionPolciy %s from Pagerduty", escalationPolicyID))
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
		r.UpdateStatus(&kubeService, fmt.Errorf("Failed to create pagerduty service"))
		return ctrl.Result{}, err
	}
	kubeService.Status.ServiceID = pdService.ID
	kubeService.Status.ServiceName = pdService.Name

	r.reconcileRoutingRules(&kubeService)

	err = r.Update(ctx, kubeService.DeepCopyObject())
	r.UpdateStatus(&kubeService, err)
	return ctrl.Result{}, err
}

func (r *PagerdutyServiceReconciler) reconcileRoutingRules(kubeService *v1.PagerdutyService) error {
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

func (r *PagerdutyServiceReconciler) destroyPagerdutyResources(kubeService *v1.PagerdutyService) error {
	logger.Info("Resource is marked for deletion. Cleaning up.")
	var err error

	ruleID := kubeService.Status.RuleID
	if ruleID != "" {
		err := r.PdClient.DeleteRulesetRule(r.RulesetID, ruleID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				logger.Info(fmt.Sprintf("Unable to delete rule %s but it does not exist.", ruleID))
			} else {
				return err
			}
		}
		logger.Info("Successfully deleted the routing rule")
	}

	serviceID := kubeService.Status.ServiceID
	if serviceID != "" {
		err = r.PdClient.DeleteService(kubeService.Status.ServiceID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				logger.Info(fmt.Sprintf("Tried to delete service %s but it does not exist.", serviceID))
			} else {
				return err
			}
			return err
		}
		logger.Info("Successfully deleted the pagerduty service")
	}

	return nil
}

// UpdateStatus sets the value of the service's Status.Status field to SUCCESS or ERROR
// based on the value of the supplied error. It persists this to etcd immediately.
func (r *PagerdutyServiceReconciler) UpdateStatus(service *v1.PagerdutyService, err error) {
	var status string
	if err == nil {
		status = "SUCCESS"
	} else {
		status = fmt.Sprintf("ERROR: %s", err.Error())
	}
	service.Status.Status = status
	r.Status().Update(context.Background(), service)
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

// GetEscalationPolicyID returns an escalation policy ID for the given service,
// If an EscalationPolicy is explicitly defined it will return that.
// Otherwise it will look for an EscalationPolicySecret, fetch the corresponding Secret, and attempt to look up the policy id
func (r *PagerdutyServiceReconciler) GetEscalationPolicyID(kubePdService *v1.PagerdutyService) (string, error) {

	// Use the explicit policy ID if supplied
	if kubePdService.Spec.EscalationPolicy != "" {
		return kubePdService.Spec.EscalationPolicy, nil
	}

	secretSpec := kubePdService.Spec.EscalationPolicySecret

	// Fail if a secret name was not supplied
	if secretSpec.Name == "" {
		return "", fmt.Errorf("No specified escalation policy ID or Secret")
	}

	// Fail if secret key was not supplied
	if secretSpec.Key == "" {
		return "", fmt.Errorf("No value for EscalationPolicySecret.Key")
	}

	ctx := context.Background()
	namespace := kubePdService.ObjectMeta.Namespace
	secret := corev1.Secret{}

	err := r.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretSpec.Name}, &secret)
	if err != nil {
		return "", err
	}

	if policyIDValue, ok := secret.Data[secretSpec.Key]; ok {
		return string(policyIDValue), nil
	}
	return "", fmt.Errorf("Could not find key %s in secret %s", secretSpec.Key, secretSpec.Name)

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
		For(&v1.PagerdutyService{}).
		Complete(r)
}

// PagerdutyInterface allows us to write a fake client for testing
// This can be replaces with pdhelpers.ServiceClient once refactors are complete
type ServiceReconcilerPagerdutyInterface interface {
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
