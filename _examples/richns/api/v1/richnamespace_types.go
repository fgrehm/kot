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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RichNamespaceSpec struct {
	ImagePullSecretRef *SecretRef        `json:"imagePullSecretsRef,omitempty"`
	CopySecrets        []SecretRef       `json:"copySecrets,omitempty"`
	DefaultResources   *DefaultResources `json:"defaultResources,omitempty"`
}

type SecretRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type DefaultResources struct {
	Request *Resources `json:"request,omitempty"`
	Limit   *Resources `json:"limit,omitempty"`
}

type Resources struct {
	CPU    *resource.Quantity `json:"cpu,omitempty"`
	Memory *resource.Quantity `json:"memory,omitempty"`
}

// RichNamespaceStatus defines the observed state of RichNamespace
type RichNamespaceStatus struct {
	ImagePullSecretRef *SecretRef        `json:"imagePullSecretsRef,omitempty"`
	CopySecrets        []SecretRef       `json:"copySecrets,omitempty"`
	DefaultResources   *DefaultResources `json:"defaultResources,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster,shortName=rns
//+kubebuilder:subresource:status

// RichNamespace is the Schema for the richnamespaces API
type RichNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RichNamespaceSpec   `json:"spec,omitempty"`
	Status RichNamespaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RichNamespaceList contains a list of RichNamespace
type RichNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RichNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RichNamespace{}, &RichNamespaceList{})
}
