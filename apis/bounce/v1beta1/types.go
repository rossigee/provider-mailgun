/*
Copyright 2025 The Crossplane Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// BounceParameters are the configurable fields of a Bounce.
type BounceParameters struct {
	// Address is the email address to add to the bounce list
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	Address string `json:"address"`

	// Code is the error code associated with the bounce (e.g., "550")
	// +kubebuilder:validation:Optional
	Code *string `json:"code,omitempty"`

	// Error is the error message describing the bounce reason
	// +kubebuilder:validation:Optional
	Error *string `json:"error,omitempty"`

	// DomainRef references the Domain resource this bounce belongs to
	// +kubebuilder:validation:Required
	DomainRef xpv1.Reference `json:"domainRef"`

	// DomainSelector selects a reference to a Domain resource
	// +kubebuilder:validation:Optional
	DomainSelector *xpv1.Selector `json:"domainSelector,omitempty"`
}

// BounceObservation are the observable fields of a Bounce.
type BounceObservation struct {
	// CreatedAt is when the bounce was recorded
	CreatedAt *string `json:"createdAt,omitempty"`
}

// A BounceSpec defines the desired state of a Bounce.
type BounceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       BounceParameters `json:"forProvider"`
}

// A BounceStatus represents the observed state of a Bounce.
type BounceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          BounceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,mailgun}
//
// This is the Crossplane v2 namespaced version.
// A Bounce is a managed resource that represents a Mailgun bounce suppression entry.
type Bounce struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BounceSpec   `json:"spec"`
	Status BounceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BounceList contains a list of Bounce
type BounceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bounce `json:"items"`
}
