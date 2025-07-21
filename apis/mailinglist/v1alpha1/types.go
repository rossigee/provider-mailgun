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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// MailingListParameters define the desired state of a Mailgun MailingList
type MailingListParameters struct {
	// Address is the mailing list email address
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	Address string `json:"address"`

	// Name is the display name for the mailing list
	Name *string `json:"name,omitempty"`

	// Description provides additional information about the mailing list
	Description *string `json:"description,omitempty"`

	// AccessLevel controls who can post to the list
	// +kubebuilder:validation:Enum=readonly;members;everyone
	// +kubebuilder:default="readonly"
	AccessLevel *string `json:"accessLevel,omitempty"`

	// ReplyPreference controls how replies are handled
	// +kubebuilder:validation:Enum=list;sender
	// +kubebuilder:default="list"
	ReplyPreference *string `json:"replyPreference,omitempty"`

	// Members is a list of email addresses to add to the mailing list
	Members []MailingListMember `json:"members,omitempty"`
}

// MailingListMember represents a member of the mailing list
type MailingListMember struct {
	// Address is the member's email address
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	Address string `json:"address"`

	// Name is the member's display name
	Name *string `json:"name,omitempty"`

	// Vars are custom variables for the member
	Vars map[string]string `json:"vars,omitempty"`

	// Subscribed indicates if the member is subscribed
	// +kubebuilder:default=true
	Subscribed *bool `json:"subscribed,omitempty"`
}

// MailingListObservation reflects the observed state of a Mailgun MailingList
type MailingListObservation struct {
	// Address is the mailing list email address
	Address string `json:"address,omitempty"`

	// Name is the display name for the mailing list
	Name string `json:"name,omitempty"`

	// Description provides additional information about the mailing list
	Description string `json:"description,omitempty"`

	// AccessLevel controls who can post to the list
	AccessLevel string `json:"accessLevel,omitempty"`

	// ReplyPreference controls how replies are handled
	ReplyPreference string `json:"replyPreference,omitempty"`

	// CreatedAt is when the mailing list was created
	CreatedAt string `json:"createdAt,omitempty"`

	// MembersCount is the number of members in the list
	MembersCount int `json:"membersCount,omitempty"`
}

// A MailingListSpec defines the desired state of a MailingList.
type MailingListSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       MailingListParameters `json:"forProvider"`
}

// A MailingListStatus represents the observed state of a MailingList.
type MailingListStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          MailingListObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A MailingList is a managed resource that represents a Mailgun MailingList
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,mailgun}
type MailingList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MailingListSpec   `json:"spec"`
	Status MailingListStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MailingListList contains a list of MailingList
type MailingListList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MailingList `json:"items"`
}
