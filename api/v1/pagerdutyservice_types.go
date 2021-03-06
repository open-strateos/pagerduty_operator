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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type LabelSpec struct {
	Key string `json:"key"`

	Value string `json:"value"`
}

// EscalationPolicySecretSpec allows you to retrieve the escalation policy from a secret
// in the same namespace as the PagerdutyService
type EscalationPolicySecretSpec struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// PagerdutyServiceSpec defines the desired state of PagerdutyService
type PagerdutyServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Description            string                     `json:"description,omitempty"`
	EscalationPolicy       string                     `json:"escalationPolicy,omitEmpty"`
	EscalationPolicySecret EscalationPolicySecretSpec `json:"escalationPolicySecret,omitEmpty"`

	// +kubebuilder:validation:MinItems:=1
	MatchLabels []LabelSpec `json:"matchLabels"`
}

// PagerdutyServiceStatus defines the observed state of PagerdutyService
type PagerdutyServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	ServiceID   string `json:"pagerdutyServiceID,omitempty"`
	ServiceName string `json:"pagerdutyServiceName,omitempty"`
	RuleID      string `json:"ruleID,omitempty"`
	Status      string `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PagerdutyService is the Schema for the pagerdutyservices API
type PagerdutyService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PagerdutyServiceSpec   `json:"spec,omitempty"`
	Status PagerdutyServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PagerdutyServiceList contains a list of PagerdutyService
type PagerdutyServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PagerdutyService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PagerdutyService{}, &PagerdutyServiceList{})
}
