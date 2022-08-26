/*
Copyright 2022.

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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OrgNamespaceSpec defines the desired state of OrgNamespace
type OrgNamespaceSpec struct {
	ImportSecrets    []SecretRef       `json:"importSecrets,omitempty"`
	DefaultResources *DefaultResources `json:"defaultResources,omitempty"`
}

type SecretRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	PullCreds bool   `json:"pullCreds"`
}

type DefaultResources struct {
	Request *Resources `json:"request,omitempty"`
	Limit   *Resources `json:"limit,omitempty"`
}

type Resources struct {
	CPU    *resource.Quantity `json:"cpu,omitempty"`
	Memory *resource.Quantity `json:"memory,omitempty"`
}

// OrgNamespaceStatus defines the observed state of OrgNamespace
type OrgNamespaceStatus struct {
	corev1.NamespaceStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster,shortName=orgns
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// OrgNamespace is the Schema for the orgnamespaces API
type OrgNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrgNamespaceSpec   `json:"spec,omitempty"`
	Status OrgNamespaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OrgNamespaceList contains a list of OrgNamespace
type OrgNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrgNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OrgNamespace{}, &OrgNamespaceList{})
}
