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

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

// k8sExternalDNSHelper is the production implementation of the
// Domain-annotation / DNS-records-ConfigMap side effects. It uses a
// controller-runtime client to read the current Domain object, mutate
// its annotations, and write it back.
//
// We deliberately re-read the Domain rather than mutating `cr` in place so
// we can apply patch semantics (Server-Side Apply) which lets the
// provider run alongside other actors writing the same object without
// fighting for ownership of unrelated fields.
type k8sExternalDNSHelper struct {
	client client.Client
}

// newK8sExternalDNSHelper returns a helper backed by the supplied k8s
// client. Pass nil to disable external-dns / ConfigMap side effects (used
// in unit tests that don't bring up an apiserver).
func newK8sExternalDNSHelper(c client.Client) *k8sExternalDNSHelper {
	return &k8sExternalDNSHelper{client: c}
}

// setExternalDNSHostname writes the
// external-dns.alpha.kubernetes.io/hostname annotation onto the Domain
// so external-dns picks it up as a managed DNS target. Opt-out is via
// `mailgun.crossplane.io/disable-external-dns: "true"` on the Domain.
//
// We set the annotation regardless of DNSVerified state — external-dns
// will be a no-op until the records exist, and writing the annotation
// when the user hasn't configured external-dns is harmless (the annotation
// is just metadata). This keeps the contract simple: opt-out means "do
// not write the annotation"; absence of opt-out means "write it once".
func (h *k8sExternalDNSHelper) setExternalDNSHostname(ctx context.Context, cr *v1beta1.Domain, domainName string) error {
	if cr == nil || h.client == nil {
		return nil
	}
	if cr.GetAnnotations()[v1beta1.AnnotationDisableExternalDNS] == "true" {
		return nil
	}

	existing := &v1beta1.Domain{}
	key := types.NamespacedName{Name: cr.GetName(), Namespace: cr.GetNamespace()}
	if err := h.client.Get(ctx, key, existing); err != nil {
		return err
	}
	annotations := existing.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[v1beta1.AnnotationExternalDNSHostname] = domainName
	existing.SetAnnotations(annotations)
	return h.client.Update(ctx, existing)
}

// ensureDNSRecordsConfigMap creates or updates the ConfigMap that carries
// the DNS records required for Mailgun verification. The ConfigMap is
// owned by the Domain (OwnerReference) so it is automatically deleted
// when the Domain is deleted.
//
// Opt-in: only writes when the user has set
// `mailgun.crossplane.io/dns-configmap: "true"` on the Domain. This is
// the default-off stance because ConfigMap write permission requires a
// ClusterRole/Binding that not every operator wants to grant.
func (h *k8sExternalDNSHelper) ensureDNSRecordsConfigMap(ctx context.Context, cr *v1beta1.Domain, records []clients.DNSRecord) error {
	if cr == nil || h.client == nil {
		return nil
	}
	if cr.GetAnnotations()[v1beta1.AnnotationConfigMapOutputEnabled] != "true" {
		return nil
	}

	cm := BuildDNSRecordsConfigMap(cr, records)
	existing := cm.DeepCopy()
	err := h.client.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, existing)
	if err != nil {
		// Create if not found.
		return h.client.Create(ctx, cm)
	}
	// Update in place; the OwnerReference will be a no-op if already set,
	// and the labels/annotations will be overwritten with the canonical
	// values from the controller so they stay in sync.
	existing.Data = cm.Data
	existing.Labels = cm.Labels
	existing.Annotations = cm.Annotations
	return h.client.Update(ctx, existing)
}
