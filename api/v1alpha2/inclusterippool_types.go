/*
Copyright 2022 Deutsche Telekom AG.

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InClusterIPPoolSpec defines the desired state of InClusterIPPool.
type InClusterIPPoolSpec struct {
	// Addresses is a list of IP addresses that can be assigned. This set of
	// addresses can be non-contiguous.
	Addresses []string `json:"addresses"`

	// Prefix is the network prefix to use.
	// +kubebuilder:validation:Maximum=128
	Prefix int `json:"prefix"`

	// Gateway
	// +optional
	Gateway string `json:"gateway,omitempty"`
}

// InClusterIPPoolStatus defines the observed state of InClusterIPPool.
type InClusterIPPoolStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Addresses",type="string",JSONPath=".spec.addresses",description="List of addresses, to allocate from"

// InClusterIPPool is the Schema for the inclusterippools API.
type InClusterIPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InClusterIPPoolSpec   `json:"spec,omitempty"`
	Status InClusterIPPoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InClusterIPPoolList contains a list of InClusterIPPool.
type InClusterIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InClusterIPPool `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Addresses",type="string",JSONPath=".spec.addresses",description="List of addresses, to allocate from"

// GlobalInClusterIPPool is the Schema for the global inclusterippools API.
// This pool type is cluster scoped. IPAddressClaims can reference
// pools of this type from any any namespace.
type GlobalInClusterIPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InClusterIPPoolSpec   `json:"spec,omitempty"`
	Status InClusterIPPoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GlobalInClusterIPPoolList contains a list of GlobalInClusterIPPool.
type GlobalInClusterIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalInClusterIPPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&InClusterIPPool{},
		&InClusterIPPoolList{},
		&GlobalInClusterIPPool{},
		&GlobalInClusterIPPoolList{},
	)
}

// PoolSpec implements the genericInClusterPool interface.
func (p *InClusterIPPool) PoolSpec() *InClusterIPPoolSpec {
	return &p.Spec
}

// PoolSpec implements the genericInClusterPool interface.
func (p *GlobalInClusterIPPool) PoolSpec() *InClusterIPPoolSpec {
	return &p.Spec
}
