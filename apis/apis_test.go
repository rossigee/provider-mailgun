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

package apis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAddToScheme(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(s))
	require.NoError(t, AddToScheme(s))

	t.Run("Domain types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "domain.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "Domain",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "Domain should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "domain.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "DomainList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "DomainList should be registered")
	})

	t.Run("MailingList types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "mailinglist.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "MailingList",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "MailingList should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "mailinglist.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "MailingListList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "MailingListList should be registered")
	})

	t.Run("Route types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "route.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "Route",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "Route should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "route.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "RouteList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "RouteList should be registered")
	})

	t.Run("Webhook types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "webhook.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "Webhook",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "Webhook should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "webhook.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "WebhookList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "WebhookList should be registered")
	})

	t.Run("Template types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "template.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "Template",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "Template should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "template.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "TemplateList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "TemplateList should be registered")
	})

	t.Run("SMTPCredential types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "smtpcredential.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "SMTPCredential",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "SMTPCredential should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "smtpcredential.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "SMTPCredentialList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "SMTPCredentialList should be registered")
	})

	t.Run("Bounce types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "bounce.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "Bounce",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "Bounce should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "bounce.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "BounceList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "BounceList should be registered")
	})

	t.Run("Unsubscribe types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "unsubscribe.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "Unsubscribe",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "Unsubscribe should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "unsubscribe.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "UnsubscribeList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "UnsubscribeList should be registered")
	})

	t.Run("Complaint types registered", func(t *testing.T) {
		gvk := schema.GroupVersionKind{
			Group:   "complaint.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "Complaint",
		}
		_, err := s.New(gvk)
		assert.NoError(t, err, "Complaint should be registered")

		listGVK := schema.GroupVersionKind{
			Group:   "complaint.mailgun.m.crossplane.io",
			Version: "v1beta1",
			Kind:    "ComplaintList",
		}
		_, err = s.New(listGVK)
		assert.NoError(t, err, "ComplaintList should be registered")
	})
}
