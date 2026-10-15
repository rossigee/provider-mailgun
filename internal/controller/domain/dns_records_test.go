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
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	clientdns "github.com/rossigee/provider-mailgun/internal/clients"
)

func TestDNSRecordsConfigMapName(t *testing.T) {
	cases := map[string]struct {
		in   string
		want string
	}{
		"PlainDomain":         {in: "example.com", want: "example.com-dns-records"},
		"MixedCaseLowercased": {in: "Example.COM", want: "example.com-dns-records"},
		"Subdomain":           {in: "mail.example.com", want: "mail.example.com-dns-records"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := dnsRecordsConfigMapName(tc.in)
			if got != tc.want {
				t.Fatalf("dnsRecordsConfigMapName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildDNSRecordsConfigMap_HasAllThreeFormats(t *testing.T) {
	cr := &v1beta1.Domain{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "mail-prod",
			UID:       "11111111-2222-3333-4444-555555555555",
		},
		Spec: v1beta1.DomainSpec{
			ForProvider: v1beta1.DomainParameters{Name: "example.com"},
		},
	}
	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org", Priority: ptrStr("10")},
		{Name: "smtp._domainkey.example.com", Type: "TXT", Value: "k=rsa; p=MIGf..."},
		{Name: "email.example.com", Type: "CNAME", Value: "mailgun.org"},
	}

	cm := BuildDNSRecordsConfigMap(cr, records)
	if cm == nil {
		t.Fatal("BuildDNSRecordsConfigMap returned nil")
	}
	if cm.Name != "example.com-dns-records" {
		t.Errorf("Name = %q, want %q", cm.Name, "example.com-dns-records")
	}
	if cm.Namespace != "mail-prod" {
		t.Errorf("Namespace = %q, want %q", cm.Namespace, "mail-prod")
	}
	if _, ok := cm.Data[dnsRecordsConfigMapKey]; !ok {
		t.Errorf("missing %s key in ConfigMap data", dnsRecordsConfigMapKey)
	}
	if _, ok := cm.Data["terraform.tf"]; !ok {
		t.Error("missing terraform.tf key in ConfigMap data")
	}
	if _, ok := cm.Data["bind-zone.txt"]; !ok {
		t.Error("missing bind-zone.txt key in ConfigMap data")
	}
	if len(cm.OwnerReferences) != 1 {
		t.Fatalf("OwnerReferences count = %d, want 1", len(cm.OwnerReferences))
	}
	if cm.OwnerReferences[0].UID != cr.GetUID() {
		t.Errorf("OwnerReferences UID = %q, want %q", cm.OwnerReferences[0].UID, cr.GetUID())
	}
}

func TestRenderRecordsYAML_IncludesAllRecords(t *testing.T) {
	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org", Priority: ptrStr("10")},
		{Name: "example.com", Type: "TXT", Value: "v=spf1 include:mailgun.org ~all"},
	}
	out := renderRecordsYAML(records)
	for _, want := range []string{"name:", "type:", "value:", "priority:", "mxa.mailgun.org", "v=spf1"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderRecordsYAML output missing %q\n---\n%s\n---", want, out)
		}
	}
}

func TestRenderTerraform_RoundTripsAllRecordTypes(t *testing.T) {
	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org", Priority: ptrStr("10")},
		{Name: "email.example.com", Type: "CNAME", Value: "mailgun.org"},
		{Name: "example.com", Type: "TXT", Value: "v=spf1 include:mailgun.org ~all"},
		{Name: "smtp._domainkey.example.com", Type: "TXT", Value: "k=rsa; p=MIGf"},
		{Name: "tracking.example.com", Type: "A", Value: "192.0.2.1"},
		{Name: "weird.example.com", Type: "SRV", Value: "0 0 0 0"},
	}
	out := renderTerraform(records)
	for _, want := range []string{
		"dns_mx_record_set", "dns_cname_record_set", "dns_txt_record_set", "dns_a_record_set",
		"unsupported record type",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("renderTerraform output missing %q", want)
		}
	}
	if !strings.Contains(out, `"10 mxa.mailgun.org."`) {
		t.Errorf("renderTerraform MX record missing trailing dot on target\n---\n%s\n---", out)
	}
	if strings.Contains(out, `"10 mxa.mailgun.org.."`) {
		t.Errorf("renderTerraform double-appended trailing dot to MX target\n---\n%s\n---", out)
	}
}

func TestRenderBindZone_HandlesMXAndTXT(t *testing.T) {
	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org", Priority: ptrStr("10")},
		{Name: "example.com", Type: "TXT", Value: "v=spf1"},
		{Name: "", Type: "TXT", Value: "v=spf1"},
	}
	out := renderBindZone(records)
	if !strings.Contains(out, "mxa.mailgun.org.") {
		t.Errorf("renderBindZone did not append trailing dot to MX target\n---\n%s\n---", out)
	}
	if strings.Contains(out, "mxa.mailgun.org..") {
		t.Errorf("renderBindZone double-appended trailing dot to MX target\n---\n%s\n---", out)
	}
	if !strings.Contains(out, "IN MX  10") {
		t.Errorf("renderBindZone missing MX line with priority\n---\n%s\n---", out)
	}
	if !strings.Contains(out, "v=spf1") {
		t.Errorf("renderBindZone missing TXT value\n---\n%s\n---", out)
	}
	if !strings.Contains(out, "@                              IN") {
		t.Errorf("renderBindZone missing '@' owner for record with empty name\n---\n%s\n---", out)
	}
}

func TestDNSRecordAnnotationKey_StableForSameName(t *testing.T) {
	r := clientdns.DNSRecord{Name: "smtp._domainkey.example.com", Type: "TXT"}
	k1 := dnsRecordAnnotationKey(r)
	k2 := dnsRecordAnnotationKey(r)
	if k1 != k2 {
		t.Errorf("annotation key not stable: %q vs %q", k1, k2)
	}
	if !strings.HasPrefix(k1, v1beta1.PrefixDNSRecordAnnotation) {
		t.Errorf("annotation key %q missing prefix %q", k1, v1beta1.PrefixDNSRecordAnnotation)
	}
}

func TestDNSRecordAnnotationKey_DiffersAcrossTypes(t *testing.T) {
	r := clientdns.DNSRecord{Name: "example.com", Type: "MX"}
	mxKey := dnsRecordAnnotationKey(r)
	r.Type = "TXT"
	txtKey := dnsRecordAnnotationKey(r)
	if mxKey == txtKey {
		t.Errorf("expected MX and TXT annotation keys to differ; got %q both", mxKey)
	}
}

func TestDNSRecordAnnotations_RoundTripsValues(t *testing.T) {
	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
		{Name: "example.com", Type: "TXT", Value: "v=spf1"},
	}
	ann := DNSRecordAnnotations(records)
	if len(ann) != 2 {
		t.Fatalf("got %d annotations, want 2; ann=%v", len(ann), ann)
	}
	for k, v := range ann {
		if v == "" {
			t.Errorf("annotation %q has empty value", k)
		}
	}
}

func TestFormatAnnotationBlock_SortsKeys(t *testing.T) {
	in := map[string]string{
		"b-key": "b-value",
		"a-key": "a-value",
		"c-key": "c-value",
	}
	out := formatAnnotationBlock(in)
	idxA := strings.Index(out, "a-key=a-value")
	idxB := strings.Index(out, "b-key=b-value")
	idxC := strings.Index(out, "c-key=c-value")
	if idxA < 0 || idxB < 0 || idxC < 0 {
		t.Fatalf("output missing expected keys:\n%s", out)
	}
	if idxA >= idxB || idxB >= idxC {
		t.Errorf("output not sorted: a=%d b=%d c=%d\n%s", idxA, idxB, idxC, out)
	}
}

func TestFormatAnnotationBlock_EmptyReturnsHint(t *testing.T) {
	out := formatAnnotationBlock(nil)
	if !strings.Contains(out, "no annotations") {
		t.Errorf("empty map should produce hint; got %q", out)
	}
}

func ptrStr(s string) *string { return &s }
