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
	//core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FactorioServerSpec defines the desired state of FactorioServer
type FactorioServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of FactorioServer. Edit FactorioServer_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	// +optional
	EnableBackups bool `json:"enableBackups"`

	// expose the entire pod spec to accelerate dev
	// This doesn't work because of: https://github.com/kubernetes/kubernetes/issues/91395
	//PodSpec core.PodTemplateSpec `json:"podSpec"`
}

// FactorioServerStatus defines the observed state of FactorioServer
type FactorioServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Port is the port the server is listening on
	// +optional
	Port int32 `json:"port,omitempty"`

	// Online mean's the server is set up and ready for work
	Online string `json:"online"`

	// Active means users are connected and playing
	Active string `json:"active"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=fs
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".status.port"
// +kubebuilder:subresource:status

// FactorioServer is the Schema for the factorioservers API
type FactorioServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FactorioServerSpec   `json:"spec,omitempty"`
	Status FactorioServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FactorioServerList contains a list of FactorioServer
type FactorioServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FactorioServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FactorioServer{}, &FactorioServerList{})
}
