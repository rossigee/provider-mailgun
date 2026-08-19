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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/miekg/dns"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	clientdns "github.com/rossigee/provider-mailgun/internal/clients"
)

// DNSProber is the abstract interface used by the controller to look up
// DNS records against authoritative resolvers. The interface exists so
// the controller can be unit-tested without touching real DNS: tests
// provide a fake prober whose returned answers are deterministic.
//
// The prober is intentionally narrow: one method, returning a slice of
// RR strings. The controller decides whether the answer matches
// Mailgun's expected value and emits the corresponding event.
type DNSProber interface {
	Lookup(ctx context.Context, name string, qtype uint16) ([]string, error)
}

// DNSProbeResult is the outcome of probing one record against the live
// DNS. ProbedValue is the joined TXT-style value (semicolon-separated
// for TXT, period-terminated FQDN for MX / CNAME, dotted-quad for A).
type DNSProbeResult struct {
	Record        clientdns.DNSRecord
	Probed        bool
	ProbedValue   string
	Matched       bool
	Err           error
	NeedsSPFMerge bool
}

// ProbeAllRecords resolves each record via the supplied prober and
// classifies the result. Records that match Mailgun's expected value
// emit a DNSRecordMatches event; mismatches emit DNSValueMismatch;
// records not found at all emit DNSNotPropagated. SPF records (TXT
// records at the domain apex whose value starts with "v=spf1") get
// additional analysis: if the probed value contains an
// `include:mailgun.org` directive but does not start with `v=spf1`,
// we flag it as a probable merge-in-progress via SPFNeedsMerge.
//
// Mailgun-specific gotchas that the prober must handle:
//
//   - Mailgun returns MX hostnames WITHOUT a trailing dot; the prober
//     compares against the expected value verbatim, so a probe of
//     `mxa.mailgun.org.` (terminated) will not match `mxa.mailgun.org`
//     even though both resolve to the same place. We normalise by
//     trimming trailing dots from both sides before comparing.
//   - TXT record values are returned as a single string (RFC 1035
//     permits long TXT records to be split into multiple <character-
//     string>s, but in practice every provider concatenates them into
//     one logical value). Mailgun returns the logical value, so a
//     direct equality check is sufficient.
//   - CNAME probes return the target as the answer; equality is
//     sufficient after the trailing-dot trim.
func ProbeAllRecords(ctx context.Context, prober DNSProber, records []clientdns.DNSRecord) []DNSProbeResult {
	out := make([]DNSProbeResult, len(records))
	for i, r := range records {
		out[i] = probeOne(ctx, prober, r)
	}
	return out
}

func probeOne(ctx context.Context, prober DNSProber, r clientdns.DNSRecord) DNSProbeResult {
	qtype := dns.StringToType[strings.ToUpper(r.Type)]
	if qtype == 0 {
		return DNSProbeResult{
			Record: r,
			Err:    fmt.Errorf("unsupported record type %q", r.Type),
		}
	}
	answers, err := prober.Lookup(ctx, dns.Fqdn(r.Name), qtype)
	if err != nil {
		return DNSProbeResult{Record: r, Err: err}
	}
	if len(answers) == 0 {
		// NXDOMAIN or NOERROR-with-no-data. We did probe successfully
		// but found nothing — distinguish from "lookup failed" by
		// returning Probed=false and no error so the caller can emit
		// DNSNotPropagated rather than DNSProbeError.
		return DNSProbeResult{Record: r}
	}

	// Normalise answer values for display purposes: trim the trailing
	// dot that anycast resolvers sometimes append to FQDN answers.
	displayed := make([]string, len(answers))
	for i, ans := range answers {
		displayed[i] = strings.TrimSuffix(ans, ".")
	}
	probed := strings.Join(displayed, "; ")

	for _, ans := range answers {
		if matchesExpected(ans, r.Value, r.Type) {
			return DNSProbeResult{
				Record:      r,
				Probed:      true,
				ProbedValue: probed,
				Matched:     true,
			}
		}
	}

	// Special case: SPF. Mailgun's expected SPF value is the bare
	// `v=spf1 include:mailgun.org ~all` directive. If the probed TXT
	// record also starts with v=spf1 and contains `include:mailgun.org`
	// but is not exactly equal to Mailgun's expected value, the user
	// almost certainly has an existing SPF they need to merge — flag it
	// so they can see the merge target. We restrict the check to TXT
	// records because that's the only RR type SPF lives in.
	if r.Type == "TXT" {
		for _, ans := range answers {
			trimmed := strings.TrimSuffix(ans, ".")
			if strings.HasPrefix(trimmed, "v=spf1") &&
				strings.Contains(trimmed, "include:mailgun.org") &&
				trimmed != r.Value {
				return DNSProbeResult{
					Record:        r,
					Probed:        true,
					ProbedValue:   probed,
					NeedsSPFMerge: true,
				}
			}
		}
	}

	return DNSProbeResult{
		Record:      r,
		Probed:      true,
		ProbedValue: probed,
	}
}

// matchesExpected performs the trailing-dot-insensitive equality check
// for FQDN-style answers (MX / CNAME / NS). For TXT and A records the
// values are compared verbatim because TXT payloads are arbitrary
// strings and A records are dotted-quad IPs.
func matchesExpected(probed, expected, qtype string) bool {
	switch strings.ToUpper(qtype) {
	case "MX", "CNAME", "NS":
		return strings.TrimSuffix(probed, ".") == strings.TrimSuffix(expected, ".")
	default:
		return probed == expected
	}
}

// EmitProbeResults fires one event per probe outcome on the supplied
// Domain. The recorder is assumed non-nil; the controller's k8sClient
// is not consulted because emitting events is a pure recorder
// operation.
func EmitProbeResults(rec event.Recorder, cr *v1beta1.Domain, results []DNSProbeResult) {
	if rec == nil {
		return
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			rec.Event(cr, event.Warning(eventReasonDNSProbeError,
				fmt.Errorf("DNS probe failed for %s %s: %w", r.Record.Type, r.Record.Name, r.Err)))
		case r.Matched:
			rec.Event(cr, event.Normal(eventReasonDNSRecordMatches,
				fmt.Sprintf("%s %s matches the expected value", r.Record.Type, r.Record.Name)))
		case r.NeedsSPFMerge:
			rec.Event(cr, event.Normal(eventReasonSPFNeedsMerge,
				fmt.Sprintf("Existing SPF detected at %s: %q\nMailgun requires: %q\nMerge by adding `include:mailgun.org` before the final ~all (or -all).",
					r.Record.Name, r.ProbedValue, r.Record.Value)))
		case r.Probed:
			rec.Event(cr, event.Warning(eventReasonDNSValueMismatch,
				fmt.Errorf("%s %s expected %q but resolved to %q",
					r.Record.Type, r.Record.Name, r.Record.Value, r.ProbedValue)))
		default:
			rec.Event(cr, event.Warning(eventReasonDNSNotPropagated,
				fmt.Errorf("%s %s not yet visible in DNS; expected %q",
					r.Record.Type, r.Record.Name, r.Record.Value)))
		}
	}
}

// MiekDNSProber is the production implementation of DNSProber. It
// resolves queries against a list of upstream resolvers (defaults to
// the well-known Google, Cloudflare and Quad9 anycast resolvers) with
// a 5-second timeout. We deliberately do not use the system resolver
// because pods in clusters often run with /etc/resolv.conf pointing at
// the cluster DNS service, which is itself a forwarding proxy —
// querying it would tell us only what the cluster DNS thinks, not
// what the authoritative servers on the internet see.
type MiekDNSProber struct {
	Resolvers []string
	Timeout   time.Duration
}

// NewMiekDNSProber returns a prober using Google's, Cloudflare's, and
// Quad9's anycast resolvers. The first one to return a non-error
// answer wins; the others are not consulted.
func NewMiekDNSProber() *MiekDNSProber {
	return &MiekDNSProber{
		Resolvers: []string{
			"8.8.8.8:53",
			"1.1.1.1:53",
			"9.9.9.9:53",
		},
		Timeout: 5 * time.Second,
	}
}

// Lookup resolves the query against the first resolver that responds.
// The prober treats NXDOMAIN and NOERROR-with-empty-answer as
// "successfully looked up, but no records exist" — both map to an
// empty []string result. Other DNS errors are propagated up.
//
// We intentionally limit to one resolver to keep latency predictable;
// the anycast IPs above are anycast so the geographic latency is
// dominated by network RTT to the user's cluster, not by resolver
// selection.
func (p *MiekDNSProber) Lookup(ctx context.Context, name string, qtype uint16) ([]string, error) {
	timeout := p.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	m := new(dns.Msg)
	m.SetQuestion(name, qtype)
	c := new(dns.Client)
	c.Timeout = timeout

	resolvers := p.Resolvers
	if len(resolvers) == 0 {
		resolvers = []string{"8.8.8.8:53"}
	}

	for _, r := range resolvers {
		resp, _, err := c.Exchange(m, r)
		if err != nil {
			continue
		}
		if resp == nil {
			continue
		}
		if resp.Rcode != dns.RcodeSuccess {
			// NXDOMAIN, SERVFAIL etc. — for our purposes "no answer"
			// is the right outcome; propagate only unexpected codes.
			if resp.Rcode == dns.RcodeNameError {
				return nil, nil
			}
			return nil, fmt.Errorf("DNS response code %d for %s", resp.Rcode, name)
		}
		return renderAnswers(resp.Answer, qtype), nil
	}
	return nil, fmt.Errorf("all resolvers failed for %s %s", dns.TypeToString[qtype], name)
}

// renderAnswers collapses an RR slice into a string slice suitable for
// equality checks. TXT records concatenate all character-strings into
// one logical value (RFC 1035 §3.3.14); MX records join host+priority;
// other types use the RR's String() representation verbatim.
func renderAnswers(rrs []dns.RR, qtype uint16) []string {
	out := make([]string, 0, len(rrs))
	for _, rr := range rrs {
		switch v := rr.(type) {
		case *dns.TXT:
			out = append(out, strings.Join(v.Txt, ""))
		case *dns.MX:
			out = append(out, fmt.Sprintf("%d %s", v.Preference, strings.TrimSuffix(v.Mx, ".")))
		default:
			out = append(out, strings.TrimSuffix(rr.String(), "."))
		}
	}
	sort.Strings(out)
	return out
}
