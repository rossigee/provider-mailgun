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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	clientdns "github.com/rossigee/provider-mailgun/internal/clients"
)

func newDomain(annotations map[string]string) *v1beta1.Domain {
	return &v1beta1.Domain{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "example",
			Namespace:   "mail-prod",
			UID:         "11111111-2222-3333-4444-555555555555",
			Annotations: annotations,
		},
		Spec: v1beta1.DomainSpec{
			ForProvider: v1beta1.DomainParameters{Name: "example.com"},
		},
	}
}

func keyFor(cr *v1beta1.Domain) types.NamespacedName {
	return types.NamespacedName{Name: cr.GetName(), Namespace: cr.GetNamespace()}
}

func keyForCM(cr *v1beta1.Domain) types.NamespacedName {
	return types.NamespacedName{Name: dnsRecordsConfigMapName(cr.Spec.ForProvider.Name), Namespace: cr.GetNamespace()}
}

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme corev1: %v", err)
	}
	if err := v1beta1.AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme domain: %v", err)
	}
	return s
}

func TestSetExternalDNSHostname_WritesAnnotationWhenEnabled(t *testing.T) {
	scheme := newScheme(t)
	cr := newDomain(nil)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cr).
		Build()
	helper := newK8sExternalDNSHelper(fakeClient)

	if err := helper.setExternalDNSHostname(context.Background(), cr, "example.com"); err != nil {
		t.Fatalf("setExternalDNSHostname: %v", err)
	}
	got := &v1beta1.Domain{}
	if err := fakeClient.Get(context.Background(), keyFor(cr), got); err != nil {
		t.Fatalf("Get Domain: %v", err)
	}
	if got.GetAnnotations()[v1beta1.AnnotationExternalDNSHostname] != "example.com" {
		t.Errorf("annotation %s = %q, want %q",
			v1beta1.AnnotationExternalDNSHostname,
			got.GetAnnotations()[v1beta1.AnnotationExternalDNSHostname],
			"example.com",
		)
	}
}

func TestSetExternalDNSHostname_RespectsOptOut(t *testing.T) {
	scheme := newScheme(t)
	cr := newDomain(map[string]string{
		v1beta1.AnnotationDisableExternalDNS: "true",
	})
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cr).
		Build()
	helper := newK8sExternalDNSHelper(fakeClient)

	if err := helper.setExternalDNSHostname(context.Background(), cr, "example.com"); err != nil {
		t.Fatalf("setExternalDNSHostname: %v", err)
	}
	got := &v1beta1.Domain{}
	if err := fakeClient.Get(context.Background(), keyFor(cr), got); err != nil {
		t.Fatalf("Get Domain: %v", err)
	}
	if _, present := got.GetAnnotations()[v1beta1.AnnotationExternalDNSHostname]; present {
		t.Errorf("annotation %s should be absent when opt-out is set",
			v1beta1.AnnotationExternalDNSHostname)
	}
}

func TestEnsureDNSRecordsConfigMap_NoopWithoutAnnotation(t *testing.T) {
	scheme := newScheme(t)
	cr := newDomain(nil)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cr).
		Build()
	helper := newK8sExternalDNSHelper(fakeClient)

	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
	}
	if err := helper.ensureDNSRecordsConfigMap(context.Background(), cr, records); err != nil {
		t.Fatalf("ensureDNSRecordsConfigMap: %v", err)
	}
	cms := &corev1.ConfigMapList{}
	if err := fakeClient.List(context.Background(), cms); err != nil {
		t.Fatalf("List ConfigMaps: %v", err)
	}
	if len(cms.Items) != 0 {
		t.Errorf("expected 0 ConfigMaps without opt-in; got %d", len(cms.Items))
	}
}

func TestEnsureDNSRecordsConfigMap_CreatesWhenOptedIn(t *testing.T) {
	scheme := newScheme(t)
	cr := newDomain(map[string]string{
		v1beta1.AnnotationConfigMapOutputEnabled: "true",
	})
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cr).
		Build()
	helper := newK8sExternalDNSHelper(fakeClient)

	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
	}
	if err := helper.ensureDNSRecordsConfigMap(context.Background(), cr, records); err != nil {
		t.Fatalf("ensureDNSRecordsConfigMap: %v", err)
	}
	cm := &corev1.ConfigMap{}
	if err := fakeClient.Get(context.Background(), keyForCM(cr), cm); err != nil {
		t.Fatalf("Get ConfigMap: %v", err)
	}
	if _, ok := cm.Data[dnsRecordsConfigMapKey]; !ok {
		t.Error("ConfigMap data missing records.yaml key")
	}
	if _, ok := cm.Data["terraform.tf"]; !ok {
		t.Error("ConfigMap data missing terraform.tf key")
	}
	if _, ok := cm.Data["bind-zone.txt"]; !ok {
		t.Error("ConfigMap data missing bind-zone.txt key")
	}
	if len(cm.OwnerReferences) != 1 {
		t.Errorf("expected 1 OwnerReference; got %d", len(cm.OwnerReferences))
	}
}

func TestEnsureDNSRecordsConfigMap_UpdatesOnSecondCall(t *testing.T) {
	scheme := newScheme(t)
	cr := newDomain(map[string]string{
		v1beta1.AnnotationConfigMapOutputEnabled: "true",
	})
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cr).
		Build()
	helper := newK8sExternalDNSHelper(fakeClient)

	records := []clientdns.DNSRecord{
		{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org"},
	}
	ctx := context.Background()
	if err := helper.ensureDNSRecordsConfigMap(ctx, cr, records); err != nil {
		t.Fatalf("first ensureDNSRecordsConfigMap: %v", err)
	}
	// Update the value to simulate Mailgun returning a different MX host.
	records[0].Value = "mxb.mailgun.org"
	if err := helper.ensureDNSRecordsConfigMap(ctx, cr, records); err != nil {
		t.Fatalf("second ensureDNSRecordsConfigMap: %v", err)
	}
	cm := &corev1.ConfigMap{}
	if err := fakeClient.Get(ctx, keyForCM(cr), cm); err != nil {
		t.Fatalf("Get ConfigMap: %v", err)
	}
	if !strings.Contains(cm.Data["records.yaml"], "mxb.mailgun.org") {
		t.Errorf("records.yaml not updated; got:\n%s", cm.Data["records.yaml"])
	}
	if strings.Contains(cm.Data["records.yaml"], "mxa.mailgun.org") {
		t.Errorf("records.yaml still contains old value mxa.mailgun.org:\n%s",
			cm.Data["records.yaml"])
	}
}
