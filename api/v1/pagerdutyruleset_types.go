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

// PagerdutyRulesetSpec defines the desired state of PagerdutyRuleset
type PagerdutyRulesetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PagerdutyRuleset. Edit PagerdutyRuleset_types.go to remove/update
	CatchallService string `json:"catchallService,omitempty"`
}

// PagerdutyRulesetStatus defines the observed state of PagerdutyRuleset
type PagerdutyRulesetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	RulesetID string `json:"rulesetID,omitempty"`
	Adopted   bool   `json:"adopted"`
}

// +kubebuilder:object:root=true

// PagerdutyRuleset is the Schema for the pagerdutyrulesets API
type PagerdutyRuleset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PagerdutyRulesetSpec   `json:"spec,omitempty"`
	Status PagerdutyRulesetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PagerdutyRulesetList contains a list of PagerdutyRuleset
type PagerdutyRulesetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PagerdutyRuleset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PagerdutyRuleset{}, &PagerdutyRulesetList{})
}
