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

// TemplateParameters are the configurable fields of a Template.
type TemplateParameters struct {
	// Domain is the domain this template belongs to.
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// Name is the template name identifier.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+$`
	Name string `json:"name"`

	// Description provides a human-readable description of the template.
	// +optional
	Description *string `json:"description,omitempty"`

	// Template contains the template content for the initial version.
	// +optional
	Template *string `json:"template,omitempty"`

	// Engine specifies the template engine to use.
	// +optional
	// +kubebuilder:validation:Enum=mustache;handlebars
	// +kubebuilder:default=mustache
	Engine *string `json:"engine,omitempty"`

	// Comment for the initial version if template content is provided.
	// +optional
	Comment *string `json:"comment,omitempty"`

	// Tag for organizing templates.
	// +optional
	Tag *string `json:"tag,omitempty"`
}

// TemplateObservation are the observable fields of a Template.
type TemplateObservation struct {
	// Name is the template identifier.
	Name string `json:"name,omitempty"`

	// Description of the template.
	Description string `json:"description,omitempty"`

	// CreatedAt is when the template was created.
	CreatedAt string `json:"createdAt,omitempty"`

	// CreatedBy indicates who created the template.
	CreatedBy string `json:"createdBy,omitempty"`

	// VersionCount is the number of versions for this template.
	VersionCount int `json:"versionCount,omitempty"`

	// ActiveVersion contains information about the active version.
	ActiveVersion *TemplateVersion `json:"activeVersion,omitempty"`
}

// TemplateVersion represents a template version
type TemplateVersion struct {
	// Tag identifying the version.
	Tag string `json:"tag,omitempty"`

	// Engine used for this version.
	Engine string `json:"engine,omitempty"`

	// CreatedAt when this version was created.
	CreatedAt string `json:"createdAt,omitempty"`

	// Comment describing this version.
	Comment string `json:"comment,omitempty"`

	// Active indicates if this is the active version.
	Active bool `json:"active,omitempty"`
}

// A TemplateSpec defines the desired state of a Template.
type TemplateSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       TemplateParameters `json:"forProvider"`
}

// A TemplateStatus represents the observed state of a Template.
type TemplateStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          TemplateObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="DOMAIN",type="string",JSONPath=".spec.forProvider.domain"
// +kubebuilder:printcolumn:name="TEMPLATE",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="VERSIONS",type="integer",JSONPath=".status.atProvider.versionCount"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,mailgun}
//
// This is the Crossplane v2 namespaced version.
// A Template is a managed resource that represents a Mailgun email template.
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec"`
	Status TemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}
