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

// SMTPCredentialParameters are the configurable fields of a SMTPCredential.
type SMTPCredentialParameters struct {
	// Domain is the domain this SMTP credential belongs to.
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// Login is the SMTP username (email address).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	Login string `json:"login"`

	// Password is the SMTP password. If not provided, Mailgun will generate one.
	// +optional
	// +kubebuilder:validation:MinLength=8
	Password *string `json:"password,omitempty"`
}

// SMTPCredentialObservation are the observable fields of a SMTPCredential.
type SMTPCredentialObservation struct {
	// Login is the SMTP username.
	Login string `json:"login,omitempty"`

	// CreatedAt is when the credential was created.
	CreatedAt string `json:"createdAt,omitempty"`

	// State indicates if the credential is active.
	State string `json:"state,omitempty"`
}

// A SMTPCredentialSpec defines the desired state of a SMTPCredential.
type SMTPCredentialSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SMTPCredentialParameters `json:"forProvider"`
}

// A SMTPCredentialStatus represents the observed state of a SMTPCredential.
type SMTPCredentialStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SMTPCredentialObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="DOMAIN",type="string",JSONPath=".spec.forProvider.domain"
// +kubebuilder:printcolumn:name="LOGIN",type="string",JSONPath=".spec.forProvider.login"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,mailgun}

// A SMTPCredential is a managed resource that represents a Mailgun SMTP credential.
type SMTPCredential struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SMTPCredentialSpec   `json:"spec"`
	Status SMTPCredentialStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SMTPCredentialList contains a list of SMTPCredential
type SMTPCredentialList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SMTPCredential `json:"items"`
}
