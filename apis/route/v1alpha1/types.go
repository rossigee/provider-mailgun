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

// RouteParameters define the desired state of a Mailgun Route
type RouteParameters struct {
	// Priority determines the order in which routes are processed (0-100, lower = higher priority)
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=0
	Priority *int `json:"priority,omitempty"`

	// Description provides a human-readable description of the route
	Description *string `json:"description,omitempty"`

	// Expression defines the filter for incoming messages
	// +kubebuilder:validation:Required
	Expression string `json:"expression"`

	// Actions define what to do with messages matching the expression
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Actions []RouteAction `json:"actions"`
}

// RouteAction defines an action to take on matching messages
type RouteAction struct {
	// Type is the action type
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=forward;store;stop
	Type string `json:"type"`

	// Destination is where to forward messages (for forward action)
	// Required for forward actions
	Destination *string `json:"destination,omitempty"`
}

// RouteObservation reflects the observed state of a Mailgun Route
type RouteObservation struct {
	// ID is the route identifier in Mailgun
	ID string `json:"id,omitempty"`

	// Priority determines the order in which routes are processed
	Priority int `json:"priority,omitempty"`

	// Description provides a human-readable description of the route
	Description string `json:"description,omitempty"`

	// Expression defines the filter for incoming messages
	Expression string `json:"expression,omitempty"`

	// Actions define what to do with messages matching the expression
	Actions []RouteAction `json:"actions,omitempty"`

	// CreatedAt is when the route was created
	CreatedAt string `json:"createdAt,omitempty"`
}

// A RouteSpec defines the desired state of a Route.
type RouteSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RouteParameters `json:"forProvider"`
}

// A RouteStatus represents the observed state of a Route.
type RouteStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RouteObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Route is a managed resource that represents a Mailgun Route
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,mailgun}
type Route struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteSpec   `json:"spec"`
	Status RouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouteList contains a list of Route
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Route `json:"items"`
}
