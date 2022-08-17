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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackstageTenantConfigSpec defines the desired state of BackstageTenantConfig
type BackstageTenantConfigSpec struct {
	// BaseURL specifies the Backstage API base URL, it can be an HTTP/S
	// address.
	// See https://backstage.io/docs/features/software-catalog/software-catalog-api
	//
	// +kubebuilder:validation:Pattern="^(http|https)://.*$"
	// +required
	BaseURL string `json:"baseURL"`

	// Interval at which to check the Backstage API for updates.
	// +required
	Interval metav1.Duration `json:"interval"`
}

// TenantResourceInventory contains a list of Kubernetes resource object references
// that have been created for tenants.
type TenantResourceInventory struct {
	// Entries of Kubernetes resource object references.
	Entries []ResourceRef `json:"entries"`
}

// ResourceRef contains the information necessary to locate a resource within a cluster.
type ResourceRef struct {
	// ID is the string representation of the Kubernetes resource object's metadata,
	// in the format '<namespace>_<name>_<group>_<kind>'.
	ID string `json:"id"`

	// Version is the API version of the Kubernetes resource object's kind.
	Version string `json:"v"`
}

// BackstageTenantConfigStatus defines the observed state of BackstageTenantConfig
type BackstageTenantConfigStatus struct {
	// TeamNames are the teams discovered from the Backstage API.
	TeamNames []string `json:"teamNames,omitempty"`
	// TenantInventory is a mapping from team to the resources generated as a
	// Tenant.
	TenantInventory map[string]TenantResourceInventory `json:"tenantInventory,omitempty"`

	// LastEtag is the last recorded etag header from the upstream API.
	LastEtag string `json:"lastEtag"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BackstageTenantConfig is the Schema for the backstagetenantconfigs API
type BackstageTenantConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackstageTenantConfigSpec   `json:"spec,omitempty"`
	Status BackstageTenantConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BackstageTenantConfigList contains a list of BackstageTenantConfig
type BackstageTenantConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackstageTenantConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackstageTenantConfig{}, &BackstageTenantConfigList{})
}
