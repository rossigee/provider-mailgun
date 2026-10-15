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
	"strings"
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/crossplane/crossplane-runtime/v2/pkg/event"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	clientdns "github.com/rossigee/provider-mailgun/internal/clients"
)

// fakeRecorder captures events emitted by the controller so tests can
// assert on them. We do not use the crossplane-runtime Recorder
// interface here because the production type writes to the API server;
// for unit tests we want a deterministic in-memory sink.
type fakeRecorder struct {
	mu     sync.Mutex
	events []event.Event
}

func (r *fakeRecorder) Event(_ runtime.Object, e event.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

func (r *fakeRecorder) WithAnnotations(_ ...string) event.Recorder { return r }

func (r *fakeRecorder) Events() []event.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]event.Event, len(r.events))
	copy(out, r.events)
	return out
}

func TestPublishDNSRecordsRequired_EmitsEventAndCreatesConfigMap(t *testing.T) {
	scheme := newScheme(t)
	cr := newDomain(map[string]string{
		v1beta1.AnnotationConfigMapOutputEnabled: "true",
	})
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cr).
		Build()
	rec := &fakeRecorder{}
	e := &external{service: nil, recorder: rec, kube: fakeClient}

	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
		{Name: "example.com", Type: "TXT", Value: "v=spf1 include:mailgun.org ~all"},
	}
	e.publishDNSRecordsRequired(context.Background(), cr, records)

	events := rec.Events()
	if len(events) == 0 {
		t.Fatal("expected at least one event, got none")
	}
	var dnsReqEvent *event.Event
	for i := range events {
		if events[i].Reason == eventReasonDNSRecordsRequired {
			dnsReqEvent = &events[i]
			break
		}
	}
	if dnsReqEvent == nil {
		t.Fatalf("expected DNSRecordsRequired event; got %d events with reasons %v",
			len(events), eventReasons(events))
	}
	if !strings.Contains(dnsReqEvent.Message, "Mailgun requires") {
		t.Errorf("DNSRecordsRequired message missing expected phrasing; got %q",
			dnsReqEvent.Message)
	}
	if !strings.Contains(dnsReqEvent.Message, "mxa.mailgun.org") {
		t.Errorf("DNSRecordsRequired message missing MX value; got %q",
			dnsReqEvent.Message)
	}

	// Verify ConfigMap was created
	cms := &corev1.ConfigMapList{}
	if err := fakeClient.List(context.Background(), cms); err != nil {
		t.Fatalf("List ConfigMaps: %v", err)
	}
	if len(cms.Items) != 1 {
		t.Fatalf("expected 1 ConfigMap; got %d", len(cms.Items))
	}
	if cms.Items[0].Data["records.yaml"] == "" {
		t.Error("ConfigMap records.yaml is empty")
	}

	// Verify external-dns annotation was written
	got := &v1beta1.Domain{}
	if err := fakeClient.Get(context.Background(), keyFor(cr), got); err != nil {
		t.Fatalf("Get Domain: %v", err)
	}
	if got.GetAnnotations()[v1beta1.AnnotationExternalDNSHostname] != "example.com" {
		t.Errorf("external-dns annotation = %q; want %q",
			got.GetAnnotations()[v1beta1.AnnotationExternalDNSHostname],
			"example.com")
	}
}

func TestPublishDNSRecordsRequired_RespectsExternalDNSOptOut(t *testing.T) {
	scheme := newScheme(t)
	cr := newDomain(map[string]string{
		v1beta1.AnnotationDisableExternalDNS: "true",
	})
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cr).
		Build()
	rec := &fakeRecorder{}
	e := &external{service: nil, recorder: rec, kube: fakeClient}

	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
	}
	e.publishDNSRecordsRequired(context.Background(), cr, records)

	got := &v1beta1.Domain{}
	if err := fakeClient.Get(context.Background(), keyFor(cr), got); err != nil {
		t.Fatalf("Get Domain: %v", err)
	}
	if _, present := got.GetAnnotations()[v1beta1.AnnotationExternalDNSHostname]; present {
		t.Error("external-dns annotation should be absent when opted out")
	}
}

func TestPublishDNSRecordsRequired_NoopWithoutKubeClient(t *testing.T) {
	cr := newDomain(nil)
	rec := &fakeRecorder{}
	e := &external{service: nil, recorder: rec, kube: nil}

	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
	}
	// Should not panic even though kube is nil.
	e.publishDNSRecordsRequired(context.Background(), cr, records)

	if len(rec.Events()) == 0 {
		t.Error("expected DNSRecordsRequired event even without kube client")
	}
}

func eventReasons(events []event.Event) []event.Reason {
	out := make([]event.Reason, len(events))
	for i, e := range events {
		out[i] = e.Reason
	}
	return out
}

// Compile-time check that fakeRecorder satisfies the Recorder interface.
var _ event.Recorder = (*fakeRecorder)(nil)

// Suppress unused warnings for the imported types we use transitively.
var (
	_ = metav1.ObjectMeta{}
	_ = corev1.ConfigMap{}
)
