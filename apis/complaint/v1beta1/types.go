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
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComplaintParameters are the configurable fields of a Complaint.
type ComplaintParameters struct {
	// Address is the email address to add to the complaint list
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+.[a-zA-Z]{2,}$"
	Address string `json:"address"`

	// DomainRef references the Domain resource this complaint belongs to
	// +kubebuilder:validation:Required
	DomainRef xpv1.Reference `json:"domainRef"`

	// DomainSelector selects a reference to a Domain resource
	// +kubebuilder:validation:Optional
	DomainSelector *xpv1.Selector `json:"domainSelector,omitempty"`
}

// ComplaintObservation are the observable fields of a Complaint.
type ComplaintObservation struct {
	// CreatedAt is when the complaint was recorded
	CreatedAt *string `json:"createdAt,omitempty"`
}

// A ComplaintSpec defines the desired state of a Complaint.
type ComplaintSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              ComplaintParameters `json:"forProvider"`
}

// A ComplaintStatus represents the observed state of a Complaint.
type ComplaintStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider             ComplaintObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,mailgun}
//
// This is the Crossplane v2 namespaced version.
// A Complaint is a managed resource that represents a Mailgun complaint suppression entry.
type Complaint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplaintSpec   `json:"spec"`
	Status ComplaintStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComplaintList contains a list of Complaint
type ComplaintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Complaint `json:"items"`
}
