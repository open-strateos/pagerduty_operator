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
	"time"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "pagerduty-operator/api/v1"
	v1 "pagerduty-operator/api/v1"
)

// PagerdutyServiceReconciler reconciles a PagerdutyService object
type PagerdutyServiceReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	PdClient      *pagerduty.Client
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

	logger.Info("Fetching PagerdutyService resource")
	if err = r.Get(ctx, req.NamespacedName, &kubeService); err != nil {
		logger.Error(err, "Unable to fetch PagerdutyService")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	spec := kubeService.Spec
	status := kubeService.Status

	var escalationPolicy *pagerduty.EscalationPolicy

	escalationPolicy, err = r.PdClient.GetEscalationPolicy(spec.EscalationPolicy, &pagerduty.GetEscalationPolicyOptions{})
	if escalationPolicy == nil {
		delay := time.Second * 30
		logger.Error(err, "Can't find the escalation policy. Will retry.", "policyID", spec.EscalationPolicy, "delay", delay)
		return ctrl.Result{Requeue: true, RequeueAfter: delay}, nil
	}

	var serviceExists bool
	if status.ServiceID != "" { // Service might already exist
		pdService, _ = r.PdClient.GetService(status.ServiceID, &pagerduty.GetServiceOptions{})
		serviceExists = pdService != nil
	}
	if pdService == nil {
		pdService = &pagerduty.Service{}
	}

	pdService.Name = r.applyPrefix(kubeService.Name)
	pdService.Description = spec.Description
	pdService.EscalationPolicy = *escalationPolicy

	if serviceExists {
		pdService, err = r.PdClient.UpdateService(*pdService)
	} else {
		pdService, err = r.PdClient.CreateService(*pdService)
	}

	return ctrl.Result{}, err
}

// applyPrefix prepends the configured prefix if applicable
func (r *PagerdutyServiceReconciler) applyPrefix(name string) string {
	if r.ServicePrefix == "" {
		return name
	}
	return r.ServicePrefix + "-" + name
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
