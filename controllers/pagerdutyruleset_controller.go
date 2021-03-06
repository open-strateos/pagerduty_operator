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
		log.V(1).Info("Unable to fetch PagerdutyRuleset", "resource", req.NamespacedName)
		return ctrl.Result{}, err
	}

	// Finalizer logic
	if kubeRuleset.DeletionTimestamp.IsZero() {
		EnsureFinalizerExists(&kubeRuleset.ObjectMeta, rulesetFinalizerKey)
	} else {
		err = r.CleanupResources(&kubeRuleset)
		if err == nil {
			log.Info("Cleanup Successful")
			EnsureFinalizerRemoved(&kubeRuleset.ObjectMeta, rulesetFinalizerKey)
			err = r.Update(ctx, &kubeRuleset)
			return ctrl.Result{}, nil // not worth doing anything else, since it's about to be deleted
		} else {
			msg := fmt.Sprintf("Cleanup error: %v", err.Error())
			r.EventRecorder.Event(&kubeRuleset, "Warning", "CleanupFail", msg)
			return ctrl.Result{Requeue: true}, err
		}
	}

	var pdRuleset *pagerduty.Ruleset
	var created bool

	if kubeRuleset.Status.RulesetID == "" {
		helper := pdhelpers.RulesetHelper{RulesetClient: r.PagerDutyClient}
		pdRuleset, created, err = helper.AdoptOrCreateRuleset(kubeRuleset.Name)
		if err != nil {
			msg := fmt.Sprintf("Unable to create ruleset: %v", err.Error())
			r.EventRecorder.Event(&kubeRuleset, "Warning", "CreateRuleset", msg)
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
	} else {
		rulesetID := kubeRuleset.Status.RulesetID
		pdRuleset, _, err = r.PagerDutyClient.GetRuleset(rulesetID)
		if err != nil {
			msg := fmt.Sprintf("Unable to fetch ruleset %s", rulesetID)
			r.EventRecorder.Event(&kubeRuleset, "Warning", "FetchPDRuleset", msg)
			r.Log.V(1).Info(msg)
			return ctrl.Result{Requeue: true}, err
		}
	}

	kubeRuleset.Status.RulesetID = pdRuleset.ID
	if created {
		kubeRuleset.Status.Created = true
	}

	err = r.Client.Update(ctx, &kubeRuleset)
	if err != nil {
		r.EventRecorder.Event(&kubeRuleset, "Warning", "CreateRuleset", err.Error())
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
		return nil // no ruleset to clean up
	} else if !ruleset.Status.Created {
		return nil // leave adopted rulesets alone, for safety
	}
	return r.PagerDutyClient.DeleteRuleset(rulesetID)
}
