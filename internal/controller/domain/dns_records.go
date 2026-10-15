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

package domain

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	clientdns "github.com/rossigee/provider-mailgun/internal/clients"
)

// dnsRecordsConfigMapName returns the ConfigMap name used to publish a
// Domain's required DNS records. Mailgun domain names are valid DNS labels
// (RFC 1035) but ConfigMap names are required to be lowercase RFC 1123; we
// lowercase to be safe. The trailing "-dns-records" suffix is fixed so the
// user can grep for it across namespaces.
func dnsRecordsConfigMapName(domainName string) string {
	return strings.ToLower(domainName) + "-dns-records"
}

// dnsRecordsConfigMapKey is the stable key for the ConfigMap data. Kept
// short so users can reference it from Kustomize patches without typos.
const dnsRecordsConfigMapKey = "records.yaml"

// BuildDNSRecordsConfigMap renders a ConfigMap that carries the DNS records
// required for the given Domain in three machine-consumable formats:
//
//   - records.yaml — a YAML list of {name, type, value, priority} objects
//     consumable by anything (kubectl apply, scripts, etc).
//   - terraform.tf — HCL resource blocks for the hashicorp/dns provider
//     (https://registry.terraform.io/providers/hashicorp/dns/latest/docs).
//   - bind-zone.txt — a BIND-format zone snippet suitable for inclusion
//     in a Bind9 zone file or as-is via `nsupdate`.
//
// The ConfigMap is owned by the Domain so it is garbage-collected when
// the Domain is deleted. The caller is responsible for creating it in the
// Domain's namespace.
func BuildDNSRecordsConfigMap(domain *v1beta1.Domain, records []clientdns.DNSRecord) *corev1.ConfigMap {
	if domain == nil {
		return nil
	}

	gvk := schema.GroupVersionKind{
		Group:   v1beta1.SchemeGroupVersion.Group,
		Version: v1beta1.SchemeGroupVersion.Version,
		Kind:    "Domain",
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsRecordsConfigMapName(domain.Spec.ForProvider.Name),
			Namespace: domain.GetNamespace(),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":  "provider-mailgun",
				"mailgun.crossplane.io/domain":  domain.Spec.ForProvider.Name,
				"mailgun.crossplane.io/records": "required",
			},
			Annotations: map[string]string{
				"mailgun.crossplane.io/managed-by": "provider-mailgun",
				"mailgun.crossplane.io/generated":  "required-dns-records",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: gvk.GroupVersion().String(),
					Kind:       gvk.Kind,
					Name:       domain.GetName(),
					UID:        domain.GetUID(),
					Controller: ptrBool(true),
				},
			},
		},
		Data: map[string]string{
			dnsRecordsConfigMapKey: renderRecordsYAML(records),
			"terraform.tf":         renderTerraform(records),
			"bind-zone.txt":        renderBindZone(records),
		},
	}
	return cm
}

// ptrBool returns &b without callers having to declare a local variable.
// Used for fixed-value pointer fields in k8s API types.
func ptrBool(b bool) *bool { return &b }

// renderRecordsYAML emits one record per line as a YAML document. We avoid
// pulling in a YAML library for this: the format is fixed and trivial, and
// pulling sigs.k8s.io/yaml just to serialise a struct we control would be
// overkill for what is effectively a printf-style blob.
func renderRecordsYAML(records []clientdns.DNSRecord) string {
	var b strings.Builder
	b.WriteString("# Required DNS records for Mailgun domain verification.\n")
	b.WriteString("# Managed by provider-mailgun; do not edit by hand.\n")
	b.WriteString("records:\n")
	for _, r := range records {
		fmt.Fprintf(&b, "  - name: %q\n", r.Name)
		fmt.Fprintf(&b, "    type: %q\n", r.Type)
		fmt.Fprintf(&b, "    value: %q\n", r.Value)
		if r.Priority != nil && *r.Priority != "" {
			fmt.Fprintf(&b, "    priority: %q\n", *r.Priority)
		}
		if r.Valid != nil {
			fmt.Fprintf(&b, "    valid: %q\n", string(*r.Valid))
		}
	}
	return b.String()
}

// renderTerraform emits one `dns_a_record_set` / `dns_cname_record_set` /
// `dns_txt_record_set` / `dns_mx_record_set` resource per record. The
// hashicorp/dns provider supports all four, so a single apply covers
// every record Mailgun might ask for.
//
// Output is intentionally minimal — users typically drive the resource
// with their own provider block, so we omit that here.
func renderTerraform(records []clientdns.DNSRecord) string {
	var b strings.Builder
	b.WriteString("# Terraform configuration for required DNS records.\n")
	b.WriteString("# Use with the hashicorp/dns provider (https://registry.terraform.io/providers/hashicorp/dns).\n")
	b.WriteString("# Managed by provider-mailgun; do not edit by hand.\n\n")

	for i, r := range records {
		switch r.Type {
		case "MX":
			fmt.Fprintf(&b, "resource \"dns_mx_record_set\" \"mailgun_%d\" {\n", i)
			fmt.Fprintf(&b, "  zone = %q\n", r.Name)
			fmt.Fprintf(&b, "  name = %q\n", r.Name)
			fmt.Fprintf(&b, "  ttl  = 3600\n")
			prio := "10"
			if r.Priority != nil && *r.Priority != "" {
				prio = *r.Priority
			}
			fmt.Fprintf(&b, "  records = [\"%s %s\"]\n", prio, ensureTrailingDot(r.Value))
		case "CNAME":
			fmt.Fprintf(&b, "resource \"dns_cname_record_set\" \"mailgun_%d\" {\n", i)
			fmt.Fprintf(&b, "  zone = %q\n", r.Name)
			fmt.Fprintf(&b, "  name = %q\n", r.Name)
			fmt.Fprintf(&b, "  ttl  = 3600\n")
			fmt.Fprintf(&b, "  records = [%q]\n", r.Value)
		case "TXT":
			fmt.Fprintf(&b, "resource \"dns_txt_record_set\" \"mailgun_%d\" {\n", i)
			fmt.Fprintf(&b, "  zone = %q\n", r.Name)
			fmt.Fprintf(&b, "  name = %q\n", r.Name)
			fmt.Fprintf(&b, "  ttl  = 3600\n")
			fmt.Fprintf(&b, "  records = [%q]\n", r.Value)
		case "A":
			fmt.Fprintf(&b, "resource \"dns_a_record_set\" \"mailgun_%d\" {\n", i)
			fmt.Fprintf(&b, "  zone = %q\n", r.Name)
			fmt.Fprintf(&b, "  name = %q\n", r.Name)
			fmt.Fprintf(&b, "  ttl  = 3600\n")
			fmt.Fprintf(&b, "  records = [%q]\n", r.Value)
		default:
			fmt.Fprintf(&b, "# unsupported record type %q (name=%q) — emit manually\n", r.Type, r.Name)
		}
		fmt.Fprintf(&b, "}\n\n")
	}
	return b.String()
}

// renderBindZone emits a zone-fragment in BIND master-file format.
// Each record becomes one line; MX records render as `<name> MX <prio> <value>.`
// (Mailgun returns MX hostnames without a trailing dot, so we add it).
// TXT / A / CNAME values are used as-is because BIND treats relative TXT
// payloads as text strings (the value is wrapped in quotes by the user at
// include-time) and A / CNAME targets that are IP / relative hostname are
// deliberately left relative to the zone.
func renderBindZone(records []clientdns.DNSRecord) string {
	var b strings.Builder
	b.WriteString("; BIND zone fragment for Mailgun domain verification.\n")
	b.WriteString("; Append the contents of this file into your zone file or\n")
	b.WriteString("; feed it to `nsupdate` after configuring the right TSIG key.\n")
	b.WriteString("; Managed by provider-mailgun; do not edit by hand.\n\n")
	b.WriteString("$TTL 3600\n\n")
	for _, r := range records {
		owner := r.Name
		if owner == "" {
			owner = "@"
		}
		switch r.Type {
		case "MX":
			prio := "10"
			if r.Priority != nil && *r.Priority != "" {
				prio = *r.Priority
			}
			fmt.Fprintf(&b, "%-30s IN MX  %s %s\n", owner, prio, ensureTrailingDot(r.Value))
		default:
			fmt.Fprintf(&b, "%-30s IN %s %s\n", owner, r.Type, r.Value)
		}
	}
	return b.String()
}

// ensureTrailingDot appends "." to DNS-hostname values that are missing it.
// BIND treats relative names as being relative to the zone origin; an MX
// target of "mxa.mailgun.org" without a trailing dot would be interpreted
// as "mxa.mailgun.org.<your-domain>", which silently breaks mail delivery.
func ensureTrailingDot(v string) string {
	if v == "" {
		return v
	}
	if strings.HasSuffix(v, ".") {
		return v
	}
	return v + "."
}

// dnsRecordAnnotationKey builds a stable annotation key for a single DNS
// record. Kubernetes annotation keys are limited to 63 characters, and
// record names + types can easily exceed that (especially for DKIM records
// like smtp._domainkey.example.com), so we hash the variable parts to a
// fixed 8-char prefix.
//
// Format: dns-controller.mailgun.crossplane.io/<type>-<hash> where <hash>
// is FNV-32 hex of "<name>".
func dnsRecordAnnotationKey(r clientdns.DNSRecord) string {
	const prefix = v1beta1.PrefixDNSRecordAnnotation
	hash := fnv32Hex(r.Name)
	return fmt.Sprintf("%s%s-%s", prefix, strings.ToLower(r.Type), hash)
}

// fnv32Hex returns the FNV-32 hash of s rendered as 8 lowercase hex
// digits. Used for short, stable, collision-friendly annotation suffixes.
func fnv32Hex(s string) string {
	const offset32 uint32 = 2166136261
	const prime32 uint32 = 16777619
	h := offset32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return fmt.Sprintf("%08x", h)
}

// DNSRecordAnnotations produces a map of annotation-key -> expected-value
// for every record. Suitable for merging into an event or directly onto
// the Domain. Keys are stable across reconciles for the same record set.
func DNSRecordAnnotations(records []clientdns.DNSRecord) map[string]string {
	out := make(map[string]string, len(records))
	for _, r := range records {
		out[dnsRecordAnnotationKey(r)] = r.Value
	}
	return out
}
