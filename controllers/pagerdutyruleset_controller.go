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

	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "pagerduty-operator/api/v1"
	v1 "pagerduty-operator/api/v1"
	"pagerduty-operator/pdhelpers"
)

const rulesetFinalizerKey = "pagerdutyruleset.core.strateos.com"

type PagerdutyReconcilerOptions struct {
	CatchallService string
}

// PagerdutyRulesetReconciler reconciles a PagerdutyRuleset object
type PagerdutyRulesetReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	EventRecorder   record.EventRecorder
	PagerDutyClient pdhelpers.RulesetClient
	Options         PagerdutyReconcilerOptions
}

// +kubebuilder:rbac:groups=core.strateos.com,resources=pagerdutyrulesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.strateos.com,resources=pagerdutyrulesets/status,verbs=get;update;patch

func (r *PagerdutyRulesetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pagerdutyruleset", req.NamespacedName)

	// your logic here
	var kubeRuleset v1.PagerdutyRuleset
	err := r.Get(ctx, req.NamespacedName, &kubeRuleset)
	if err != nil {
		log.V(1).Info("Unable to fetch PagerdutyRuleset: %v", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if kubeRuleset.DeletionTimestamp.IsZero() {
		EnsureFinalizerExists(&kubeRuleset.ObjectMeta, rulesetFinalizerKey)
	} else {
		err = r.CleanupResources(&kubeRuleset)
		if err == nil {
			log.Info("Cleanup Successful")
			EnsureFinalizerRemoved(&kubeRuleset.ObjectMeta, rulesetFinalizerKey)
			err = r.Update(ctx, &kubeRuleset)
		} else {
			msg := fmt.Sprintf("Cleanup error: %v", err.Error())
			r.EventRecorder.Event(&kubeRuleset, "Warning", "CleanupFail", msg)
			log.Error(err, "Cleanup Failed", "resource", req.NamespacedName)
		}
	}

	var pdRuleset *pagerduty.Ruleset
	var created bool

	if kubeRuleset.Status.RulesetID == "" {
		opts := pdhelpers.RulesetOptions{
			Name:                &kubeRuleset.Name,
			CatchallServiceName: r.Options.CatchallService,
		}
		helper := pdhelpers.RulesetHelper{RulesetClient: r.PagerDutyClient}
		pdRuleset, created, err = helper.AdoptOrCreateRuleset(&opts)
		if err != nil {
			msg := fmt.Sprintf("Unable to create ruleset: %v", err.Error())
			r.EventRecorder.Event(&kubeRuleset, "Warning", "CreateRuleset", msg)
			r.Log.V(1).Info(msg)
			return ctrl.Result{Requeue: true}, err
		}

		var adopedOrCreated string
		if created {
			adopedOrCreated = "Created"
		} else {
			adopedOrCreated = "Adopted"
		}
		msg := fmt.Sprintf("%s ruleset %s (ID: %s)", adopedOrCreated, pdRuleset.Name, pdRuleset.ID)
		r.EventRecorder.Event(&kubeRuleset, "Normal", "CreateRuleset", msg)
		r.Log.V(1).Info(msg)
	}

	// FIXME!
	if pdRuleset == nil {
		err := fmt.Errorf("Something went very wrong. pdRuleset is nil")
		log.Error(err, "This shouldn't be able to happen, but sometimes does.")
		return ctrl.Result{Requeue: true}, err
	}

	kubeRuleset.Status.RulesetID = pdRuleset.ID
	kubeRuleset.Status.Adopted = kubeRuleset.Status.Adopted || (!created)

	err = r.Client.Update(ctx, &kubeRuleset)
	if err != nil {
		r.EventRecorder.Event(&kubeRuleset, "Warning", "CreateRuleset", err.Error())
		r.Log.V(1).Info(err.Error())
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *PagerdutyRulesetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PagerdutyRuleset{}).
		Complete(r)
}

func (r *PagerdutyRulesetReconciler) CleanupResources(ruleset *v1.PagerdutyRuleset) error {
	rulesetID := ruleset.Status.RulesetID
	if rulesetID == "" {
		return fmt.Errorf("Empty rulesetID")
	}
	return r.PagerDutyClient.DeleteRuleset(rulesetID)
}

// UpdateStatus sets the value of the service's Status.Status field to SUCCESS or ERROR
// based on the value of the supplied error. It persists this to etcd immediately.
func (r *PagerdutyRulesetReconciler) UpdateStatus(ruleset *v1.PagerdutyRuleset, err error) {
	var state string
	if err == nil {
		state = "SUCCESS"
	} else {
		state = fmt.Sprintf("ERROR: %s", err.Error())
	}
	ruleset.Status.State = state
	r.Status().Update(context.Background(), ruleset)
}
