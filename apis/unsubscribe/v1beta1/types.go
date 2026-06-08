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

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// UnsubscribeParameters are the configurable fields of an Unsubscribe.
type UnsubscribeParameters struct {
	// Address is the email address to add to the unsubscribe list
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	Address string `json:"address"`

	// Tags is a comma-separated list of tags associated with the unsubscribe
	// +kubebuilder:validation:Optional
	Tags *string `json:"tags,omitempty"`

	// DomainRef references the Domain resource this unsubscribe belongs to
	// +kubebuilder:validation:Required
	DomainRef xpv1.Reference `json:"domainRef"`

	// DomainSelector selects a reference to a Domain resource
	// +kubebuilder:validation:Optional
	DomainSelector *xpv1.Selector `json:"domainSelector,omitempty"`
}

// UnsubscribeObservation are the observable fields of an Unsubscribe.
type UnsubscribeObservation struct {
	// CreatedAt is when the unsubscribe was recorded
	CreatedAt *string `json:"createdAt,omitempty"`
}

// An UnsubscribeSpec defines the desired state of an Unsubscribe.
type UnsubscribeSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider       UnsubscribeParameters `json:"forProvider"`
}

// An UnsubscribeStatus represents the observed state of an Unsubscribe.
type UnsubscribeStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider          UnsubscribeObservation `json:"atProvider,omitempty"`
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
// An Unsubscribe is a managed resource that represents a Mailgun unsubscribe suppression entry.
type Unsubscribe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UnsubscribeSpec   `json:"spec"`
	Status UnsubscribeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UnsubscribeList contains a list of Unsubscribe
type UnsubscribeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Unsubscribe `json:"items"`
}
