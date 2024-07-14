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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SystemRule struct {
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Optional
	Id string `json:"id,omitempty"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Load;AvgRT;Concurrency;InboundQPS;CpuUsage
	// +kubebuilder:validation:Required
	MetricType string `json:"metricType"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=NoAdaptive;BBR
	// +kubebuilder:validation:Required
	Strategy string `json:"strategy"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	TriggerCount int64 `json:"triggerCount"`
}

// SystemRulesSpec defines the desired state of SystemRules
type SystemRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Type=array
	// +kubebuilder:validation:Optional
	Rules []SystemRule `json:"rules"`
}

// SystemRulesStatus defines the observed state of SystemRules
type SystemRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// SystemRules is the Schema for the systemrules API
type SystemRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SystemRulesSpec   `json:"spec,omitempty"`
	Status SystemRulesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SystemRulesList contains a list of SystemRules
type SystemRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SystemRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SystemRules{}, &SystemRulesList{})
}
