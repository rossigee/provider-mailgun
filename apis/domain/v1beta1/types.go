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

// DomainParameters define the desired state of a Mailgun Domain
type DomainParameters struct {
	// Name is the domain name to create
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	Name string `json:"name"`

	// Type specifies the domain type (sending or receiving)
	// +kubebuilder:validation:Enum=sending;receiving
	// +kubebuilder:default="sending"
	Type *string `json:"type,omitempty"`

	// ForceDKIMAuthority forces DKIM authority even if subdomain
	// +kubebuilder:default=false
	ForceDKIMAuthority *bool `json:"forceDkimAuthority,omitempty"`

	// DKIMKeySize specifies the DKIM key size (1024 or 2048)
	// +kubebuilder:validation:Enum=1024;2048
	// +kubebuilder:default=1024
	DKIMKeySize *int `json:"dkimKeySize,omitempty"`

	// IPs is a list of IP addresses to whitelist for this domain
	IPs []string `json:"ips,omitempty"`

	// Tracking settings for the domain
	Tracking *DomainTracking `json:"tracking,omitempty"`

	// SMTP password for the domain (if not set, will be auto-generated)
	SMTPPassword *string `json:"smtpPassword,omitempty"`

	// Spam action for the domain
	// +kubebuilder:validation:Enum=disabled;block;tag
	// +kubebuilder:default="disabled"
	SpamAction *string `json:"spamAction,omitempty"`

	// Web scheme for tracking URLs
	// +kubebuilder:validation:Enum=http;https
	// +kubebuilder:default="http"
	WebScheme *string `json:"webScheme,omitempty"`

	// Wildcard setting for the domain
	// +kubebuilder:default=false
	Wildcard *bool `json:"wildcard,omitempty"`
}

// DomainTracking defines tracking settings for a domain
type DomainTracking struct {
	// Click tracking enabled
	// +kubebuilder:default=false
	Click *bool `json:"click,omitempty"`

	// Open tracking enabled
	// +kubebuilder:default=false
	Open *bool `json:"open,omitempty"`

	// Unsubscribe tracking enabled
	// +kubebuilder:default=false
	Unsubscribe *bool `json:"unsubscribe,omitempty"`
}

// DomainObservation reflects the observed state of a Mailgun Domain
type DomainObservation struct {
	// ID is the domain identifier in Mailgun
	ID string `json:"id,omitempty"`

	// State is the current state of the domain (active, unverified, disabled)
	State string `json:"state,omitempty"`

	// CreatedAt is when the domain was created
	CreatedAt string `json:"createdAt,omitempty"`

	// SMTPLogin is the SMTP login for the domain
	SMTPLogin string `json:"smtpLogin,omitempty"`

	// SMTPPassword is the SMTP password for the domain
	SMTPPassword string `json:"smtpPassword,omitempty"`

	// RequiredDNSRecords contains the DNS records that need to be configured
	RequiredDNSRecords []DNSRecord `json:"requiredDnsRecords,omitempty"`

	// Receiving DNS records for incoming mail
	ReceivingDNSRecords []DNSRecord `json:"receivingDnsRecords,omitempty"`

	// Sending DNS records for outgoing mail
	SendingDNSRecords []DNSRecord `json:"sendingDnsRecords,omitempty"`
}

// DNSRecord represents a DNS record required for domain configuration
type DNSRecord struct {
	// Name is the DNS record name
	Name string `json:"name,omitempty"`

	// Type is the DNS record type (TXT, CNAME, MX)
	Type string `json:"type,omitempty"`

	// Value is the DNS record value
	Value string `json:"value,omitempty"`

	// Priority is the MX record priority (for MX records)
	Priority *int `json:"priority,omitempty"`

	// Valid indicates if the DNS record is properly configured
	Valid *bool `json:"valid,omitempty"`
}

// A DomainSpec defines the desired state of a Domain.
type DomainSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DomainParameters `json:"forProvider"`
}

// A DomainStatus represents the observed state of a Domain.
type DomainStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DomainObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Domain is a namespaced managed resource that represents a Mailgun Domain.
// This is the Crossplane v2 namespaced version.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,mailgun}
type Domain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DomainSpec   `json:"spec"`
	Status DomainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DomainList contains a list of Domain
type DomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Domain `json:"items"`
}
