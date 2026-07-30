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

// DomainParameters define the desired state of a Mailgun Domain
type DomainParameters struct {
	// Name is the domain name to create
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
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

	// DNSProvider optionally wires the Domain to a third-party DNS hosting
	// provider so the controller can push the Mailgun DNS records directly
	// (e.g. Cloudflare). When unset, the controller falls back to generic
	// automation: ConfigMap output, external-dns annotation, and (opt-in)
	// local DNS propagation probing. Only one provider can be configured per
	// Domain.
	//
	// This field is optional and additive; existing Domains without a
	// DNSProvider continue to behave exactly as before.
	DNSProvider *DNSProviderSpec `json:"dnsProvider,omitempty"`
}

// DNSProviderSpec selects which third-party DNS hosting provider the
// controller should use to publish Mailgun's required DNS records.
type DNSProviderSpec struct {
	// Cloudflare configures the controller to push the DNS records to
	// Cloudflare. When set, the controller authenticates with the supplied
	// API token and writes MX / TXT / CNAME records into the specified zone.
	// The Domain is not deleted from Cloudflare automatically when the
	// Crossplane resource is deleted - set cloudflare.keepOnDelete to true
	// (the default) to preserve the records for inspection; flip to false to
	// remove them.
	Cloudflare *CloudflareProviderSpec `json:"cloudflare,omitempty"`
}

// CloudflareProviderSpec configures Cloudflare as the DNS hosting provider
// for the records required by a Mailgun Domain.
type CloudflareProviderSpec struct {
	// APITokenSecretRef references a Kubernetes Secret containing a single
	// data key 'apiToken' with the Cloudflare API token. The token must
	// have Zone.DNS edit permission for the zone identified by ZoneID.
	// +kubebuilder:validation:Required
	APITokenSecretRef *SecretKeyRef `json:"apiTokenSecretRef"`

	// ZoneID is the Cloudflare Zone ID where the records will be created.
	// Find this in the Cloudflare dashboard under the domain overview page
	// (lower-right API section).
	// +kubebuilder:validation:Required
	ZoneID string `json:"zoneID"`

	// KeepOnDelete, when true (default), leaves DNS records in Cloudflare
	// when the Domain is deleted from Crossplane. Set to false to remove
	// them. We default to true so a careless `kubectl delete` does not
	// knock your email offline.
	// +kubebuilder:default=true
	KeepOnDelete *bool `json:"keepOnDelete,omitempty"`
}

// SecretKeyRef is a reference to a single key inside a Kubernetes Secret.
type SecretKeyRef struct {
	// Name is the name of the Secret in the Domain's namespace.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Key is the data key inside the Secret. Defaults to "apiToken" when
	// omitted; other consumers can reuse this type with different defaults.
	// +kubebuilder:default="apiToken"
	Key string `json:"key,omitempty"`
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

	// DNSVerified indicates whether all required DNS records are properly configured
	DNSVerified *bool `json:"dnsVerified,omitempty"`

	// RequiredDNSRecords contains the DNS records that need to be configured
	RequiredDNSRecords []DNSRecord `json:"requiredDnsRecords,omitempty"`

	// Receiving DNS records for incoming mail
	ReceivingDNSRecords []DNSRecord `json:"receivingDnsRecords,omitempty"`

	// Sending DNS records for outgoing mail
	SendingDNSRecords []DNSRecord `json:"sendingDnsRecords,omitempty"`
}

// DNSRecord represents a DNS record required for domain configuration.
// `Valid` mirrors Mailgun's v4 API verbatim: "valid" if Mailgun confirms the
// record is properly configured, "unknown" if it has not yet been verified.
type DNSRecord struct {
	// Name is the DNS record name
	Name string `json:"name,omitempty"`

	// Type is the DNS record type (TXT, CNAME, MX, A)
	Type string `json:"type,omitempty"`

	// Value is the DNS record value
	Value string `json:"value,omitempty"`

	// Priority is the MX record priority (for MX records), as a string per
	// Mailgun's v4 API (e.g. "10")
	Priority *string `json:"priority,omitempty"`

	// Valid is the verification status returned by Mailgun.
	// +kubebuilder:validation:Enum=valid;unknown
	Valid *RecordValidity `json:"valid,omitempty"`
}

// RecordValidity is the verification status of a DNS record as reported by
// Mailgun's v4 Domains API.
type RecordValidity string

// Known RecordValidity values. Mailgun may also return an empty string for
// records that have not yet been queried; treat that as not-yet-verified.
const (
	// RecordValidityValid indicates Mailgun confirmed the DNS record is
	// properly configured.
	RecordValidityValid RecordValidity = "valid"

	// RecordValidityUnknown indicates Mailgun has not yet verified the DNS
	// record (typically because it has not been propagated yet).
	RecordValidityUnknown RecordValidity = "unknown"
)

// IsVerified returns true when the record has been verified by Mailgun.
func (v *RecordValidity) IsVerified() bool {
	return v != nil && *v == RecordValidityValid
}

// Annotation keys the provider sets on Domain resources to coordinate DNS
// automation with external tooling.
const (
	// AnnotationExternalDNSHostname is the hostname annotation read by
	// external-dns (https://github.com/kubernetes-sigs/external-dns) so it
	// picks the Domain up as a managed DNS target. Set automatically by
	// the controller unless the user opts out via
	// AnnotationDisableExternalDNS.
	AnnotationExternalDNSHostname = "external-dns.alpha.kubernetes.io/hostname"

	// AnnotationDisableExternalDNS, when set to "true", prevents the
	// controller from writing AnnotationExternalDNSHostname. Use this when
	// you have external-dns installed but do not want it to manage the
	// Mailgun DNS records (for example, because you manage them via a
	// Composition that uses provider-cloudflare).
	AnnotationDisableExternalDNS = "mailgun.crossplane.io/disable-external-dns"

	// AnnotationDNSProbeEnabled, when set to "true", makes the controller
	// query authoritative DNS resolvers for each required record and emit
	// per-record diagnostic events (DNSNotPropagated, DNSValueMismatch,
	// SPFNeedsMerge, DNSRecordMatches). Default is off so the provider
	// does not require outbound DNS egress from the cluster by default.
	AnnotationDNSProbeEnabled = "mailgun.crossplane.io/dns-probe"

	// AnnotationConfigMapOutputEnabled, when set to "true", makes the
	// controller write/update a ConfigMap named "<domain>-dns-records"
	// containing the required DNS records in records.yaml / terraform.tf /
	// bind-zone.txt formats. Default is off so the provider does not
	// require ConfigMap write permissions by default.
	AnnotationConfigMapOutputEnabled = "mailgun.crossplane.io/dns-configmap"

	// PrefixDNSRecordAnnotation is prepended to event annotations that
	// carry the expected value of a single DNS record. The full key is
	// PrefixDNSRecordAnnotation + "<type>-<name>" (truncated/hashed to
	// stay within Kubernetes' 63-character annotation key limit).
	PrefixDNSRecordAnnotation = "dns-controller.mailgun.crossplane.io/"
)

// A DomainSpec defines the desired state of a Domain.
type DomainSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              DomainParameters `json:"forProvider"`
}

// A DomainStatus represents the observed state of a Domain.
type DomainStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider             DomainObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Domain is a namespaced managed resource that represents a Mailgun Domain.
// This is the Crossplane v2 namespaced version.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="DNS-VERIFIED",type="string",JSONPath=".status.atProvider.dnsVerified"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,mailgun}
// +kubebuilder:storageversion
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
