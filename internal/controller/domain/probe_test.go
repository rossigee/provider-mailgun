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
	"context"
	"errors"
	"testing"

	"github.com/miekg/dns"

	clientdns "github.com/rossigee/provider-mailgun/internal/clients"
)

// fakeProber is a deterministic DNSProber used to drive probeOne in
// unit tests without touching real DNS. The probe map is keyed by the
// fully-qualified name (with trailing dot) that the prober sees.
type fakeProber struct {
	answers map[string][]string
	err     error
}

func (f *fakeProber) Lookup(_ context.Context, name string, _ uint16) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.answers[name], nil
}

func TestProbeOne_MXMatchAfterTrailingDotTrim(t *testing.T) {
	prober := &fakeProber{
		answers: map[string][]string{
			"example.com.": {"mxa.mailgun.org."},
		},
	}
	rec := clientdns.DNSRecord{
		Name:  "example.com",
		Type:  "MX",
		Value: "mxa.mailgun.org",
	}
	r := probeOne(context.Background(), prober, rec)
	if !r.Probed {
		t.Fatal("expected Probed=true after successful lookup")
	}
	if !r.Matched {
		t.Errorf("expected Matched=true; got %#v", r)
	}
	if r.Err != nil {
		t.Errorf("expected nil error; got %v", r.Err)
	}
}

func TestProbeOne_MXMismatchSurfacesProbedValue(t *testing.T) {
	prober := &fakeProber{
		answers: map[string][]string{
			"example.com.": {"mx.other-provider.org."},
		},
	}
	rec := clientdns.DNSRecord{
		Name:  "example.com",
		Type:  "MX",
		Value: "mxa.mailgun.org",
	}
	r := probeOne(context.Background(), prober, rec)
	if !r.Probed || r.Matched {
		t.Errorf("expected Probed=true and Matched=false; got %#v", r)
	}
	if r.ProbedValue != "mx.other-provider.org" {
		t.Errorf("ProbedValue = %q; want %q", r.ProbedValue, "mx.other-provider.org")
	}
}

func TestProbeOne_NotPropagatedOnEmptyAnswer(t *testing.T) {
	prober := &fakeProber{
		answers: map[string][]string{
			"smtp._domainkey.example.com.": {},
		},
	}
	rec := clientdns.DNSRecord{
		Name:  "smtp._domainkey.example.com",
		Type:  "TXT",
		Value: "k=rsa; p=MIGf",
	}
	r := probeOne(context.Background(), prober, rec)
	if r.Probed {
		t.Errorf("expected Probed=false when no answers returned; got %#v", r)
	}
}

func TestProbeOne_TXTExactMatch(t *testing.T) {
	prober := &fakeProber{
		answers: map[string][]string{
			"smtp._domainkey.example.com.": {"k=rsa; p=MIGf"},
		},
	}
	rec := clientdns.DNSRecord{
		Name:  "smtp._domainkey.example.com",
		Type:  "TXT",
		Value: "k=rsa; p=MIGf",
	}
	r := probeOne(context.Background(), prober, rec)
	if !r.Matched {
		t.Errorf("expected Matched=true; got %#v", r)
	}
}

func TestProbeOne_SPFNeedsMergeDetected(t *testing.T) {
	// User has existing SPF: "v=spf1 include:_spf.google.com ~all"
	// Mailgun wants:                  "v=spf1 include:mailgun.org ~all"
	// Both contain "include:mailgun.org"? No - the existing one has google.
	// So the merge detector should fire only when mailgun.org is present.
	prober := &fakeProber{
		answers: map[string][]string{
			"example.com.": {"v=spf1 include:_spf.google.com include:mailgun.org ~all"},
		},
	}
	rec := clientdns.DNSRecord{
		Name:  "example.com",
		Type:  "TXT",
		Value: "v=spf1 include:mailgun.org ~all",
	}
	r := probeOne(context.Background(), prober, rec)
	if !r.NeedsSPFMerge {
		t.Errorf("expected NeedsSPFMerge=true; got %#v", r)
	}
	if r.Matched {
		t.Errorf("NeedsSPFMerge should imply Matched=false; got %#v", r)
	}
}

func TestProbeOne_UnsupportedTypeReturnsError(t *testing.T) {
	prober := &fakeProber{}
	// "BOGUS" is not a real DNS RR type, so StringToType returns 0 and
	// the prober should refuse to query.
	rec := clientdns.DNSRecord{Name: "x.example.com", Type: "BOGUS", Value: "0 0 0 0"}
	r := probeOne(context.Background(), prober, rec)
	if r.Err == nil {
		t.Fatal("expected error for unsupported type; got nil")
	}
}

func TestProbeOne_LookupErrorPropagates(t *testing.T) {
	prober := &fakeProber{err: errors.New("resolver unreachable")}
	rec := clientdns.DNSRecord{Name: "example.com", Type: "A", Value: "192.0.2.1"}
	r := probeOne(context.Background(), prober, rec)
	if r.Err == nil {
		t.Fatal("expected error to be propagated")
	}
	if r.Probed {
		t.Errorf("Probed should be false when lookup failed; got %#v", r)
	}
}

func TestProbeAllRecords_IteratesAll(t *testing.T) {
	prober := &fakeProber{
		answers: map[string][]string{
			"example.com.":                 {"mxa.mailgun.org."},
			"smtp._domainkey.example.com.": {"k=rsa; p=ABC"},
		},
	}
	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
		{Name: "smtp._domainkey.example.com", Type: "TXT", Value: "k=rsa; p=ABC"},
	}
	results := ProbeAllRecords(context.Background(), prober, records)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for i, r := range results {
		if !r.Matched {
			t.Errorf("result[%d] (%s %s) not matched; got %#v", i, r.Record.Type, r.Record.Name, r)
		}
	}
}

func TestMiekDNSProber_ParseRcodeNameError(t *testing.T) {
	// We can't easily simulate DNS responses without a real socket, but
	// we can at least verify that the prober is constructed correctly.
	p := NewMiekDNSProber()
	if p.Timeout == 0 {
		t.Error("NewMiekDNSProber should set a non-zero default timeout")
	}
	if len(p.Resolvers) == 0 {
		t.Error("NewMiekDNSProber should provide at least one resolver")
	}
}

func TestRenderAnswers_TXTConcatenation(t *testing.T) {
	rr := &dns.TXT{
		Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
		Txt: []string{"part1", "part2"},
	}
	out := renderAnswers([]dns.RR{rr}, dns.TypeTXT)
	if len(out) != 1 || out[0] != "part1part2" {
		t.Errorf("renderAnswers TXT concat = %v; want [part1part2]", out)
	}
}

func TestRenderAnswers_MXFormat(t *testing.T) {
	rr := &dns.MX{
		Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 300},
		Preference: 10,
		Mx:         "mxa.mailgun.org.",
	}
	out := renderAnswers([]dns.RR{rr}, dns.TypeMX)
	if len(out) != 1 || out[0] != "10 mxa.mailgun.org" {
		t.Errorf("renderAnswers MX = %v; want [10 mxa.mailgun.org]", out)
	}
}

func TestMatchesExpected_TrailingDotInsensitive(t *testing.T) {
	cases := map[string]struct {
		probed, expected, qtype string
		want                    bool
	}{
		"MXMatchWithTrailingDot": {probed: "mxa.mailgun.org.", expected: "mxa.mailgun.org", qtype: "MX", want: true},
		"MXMismatch":             {probed: "mx.other.org.", expected: "mxa.mailgun.org", qtype: "MX", want: false},
		"CNAMEMatch":             {probed: "mailgun.org.", expected: "mailgun.org", qtype: "CNAME", want: true},
		"TXTExact":               {probed: "v=spf1", expected: "v=spf1", qtype: "TXT", want: true},
		"TXTMismatch":            {probed: "v=spf2", expected: "v=spf1", qtype: "TXT", want: false},
		"AMatch":                 {probed: "192.0.2.1", expected: "192.0.2.1", qtype: "A", want: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := matchesExpected(tc.probed, tc.expected, tc.qtype); got != tc.want {
				t.Errorf("matchesExpected(%q,%q,%q) = %v, want %v",
					tc.probed, tc.expected, tc.qtype, got, tc.want)
			}
		})
	}
}
