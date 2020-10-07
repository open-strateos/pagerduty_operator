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
	APIKey        string // pagerduty API key
	ServicePrefix string // append to service names
}

// +kubebuilder:rbac:groups=core.strateos.com,resources=pagerdutyservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.strateos.com,resources=pagerdutyservices/status,verbs=get;update;patch

func (r *PagerdutyServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("pagerdutyservice", req.NamespacedName)

	var service v1.PagerdutyService
	r.Get(ctx, req.NamespacedName, &service)

	return ctrl.Result{}, nil
}

func (r *PagerdutyServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PagerdutyService{}).
		Complete(r)
}
