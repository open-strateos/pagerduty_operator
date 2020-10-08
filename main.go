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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1 "pagerduty-operator/api/v1"
	"pagerduty-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var pagerdutyAPIKey string
	var servicePrefix string
	var rulesetID string

	flag.StringVar(&metricsAddr, "metrics-addr", getEnv("METRICS_ADDR", ":8080"), "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&pagerdutyAPIKey, "api-key", getEnv("PAGERDUTY_API_KEY", ""), "Authorization key for the pagerduty API.")
	flag.StringVar(&servicePrefix, "service-prefix", getEnv("PAGERDUTY_SERVICE_PREFIX", ""), "Prefix to be added to Pagerduty Service names")
	flag.StringVar(&rulesetID, "ruleset", getEnv("PAGERDUTY_RULESET_ID", ""), "ID of the ruleset to append routing rules to.")
	flag.Parse()

	if pagerdutyAPIKey == "" {
		setupLog.Info("API key is required.")
		os.Exit(1)
	}
	if rulesetID == "" {
		setupLog.Info("Ruleset ID is required")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "40e424f7.strateos.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	pdClient := pagerduty.NewClient(pagerdutyAPIKey)
	_ = getRulesetOrDie(pdClient, rulesetID)

	if err = (&controllers.PagerdutyServiceReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("PagerdutyService"),
		Scheme:        mgr.GetScheme(),
		PdClient:      pdClient,
		RulesetID:     rulesetID,
		ServicePrefix: servicePrefix,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PagerdutyService")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getRulesetOrDie(pdClient *pagerduty.Client, rulesetID string) *pagerduty.Ruleset {
	ruleset, _, err := pdClient.GetRuleset(rulesetID)
	if err != nil {
		setupLog.Error(err, fmt.Sprintf("Ruleset %s does not exist", rulesetID))
		os.Exit(1)
	}
	return ruleset
}
