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

package webhook

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	"github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
	bouncetypes "github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglisttypes "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
)

// MockWebhookClient for testing
type MockWebhookClient struct {
	webhooks map[string]*v1beta1.WebhookObservation
	err      error
}

func (m *MockWebhookClient) CreateWebhook(ctx context.Context, domain string, webhook *v1beta1.WebhookParameters) (*v1beta1.WebhookObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := &v1beta1.WebhookObservation{
		ID:        "webhook_123",
		EventType: webhook.EventType,
		URL:       webhook.URL,
		CreatedAt: "2025-01-01T00:00:00Z",
	}

	if webhook.Username != nil {
		result.Username = *webhook.Username
	}

	key := domain + "/" + webhook.EventType
	if m.webhooks == nil {
		m.webhooks = make(map[string]*v1beta1.WebhookObservation)
	}
	m.webhooks[key] = result

	return result, nil
}

func (m *MockWebhookClient) GetWebhook(ctx context.Context, domain, eventType string) (*v1beta1.WebhookObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + eventType
	if webhook, exists := m.webhooks[key]; exists {
		return webhook, nil
	}

	return nil, errors.New("webhook not found (404)")
}

func (m *MockWebhookClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *v1beta1.WebhookParameters) (*v1beta1.WebhookObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + eventType
	if existing, exists := m.webhooks[key]; exists {
		// Update modifiable fields
		existing.URL = webhook.URL
		if webhook.Username != nil {
			existing.Username = *webhook.Username
		}
		return existing, nil
	}

	return nil, errors.New("webhook not found (404)")
}

func (m *MockWebhookClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	if m.err != nil {
		return m.err
	}

	key := domain + "/" + eventType
	delete(m.webhooks, key)
	return nil
}

// Implement other required client methods as no-ops
func (m *MockWebhookClient) CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockWebhookClient) CreateMailingList(ctx context.Context, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetMailingList(ctx context.Context, address string) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) UpdateMailingList(ctx context.Context, address string, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

func (m *MockWebhookClient) CreateRoute(ctx context.Context, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetRoute(ctx context.Context, id string) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) UpdateRoute(ctx context.Context, id string, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *MockWebhookClient) CreateSMTPCredential(ctx context.Context, domain string, credential *smtpcredentialtypes.SMTPCredentialParameters) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetSMTPCredential(ctx context.Context, domain, login string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

func (m *MockWebhookClient) CreateTemplate(ctx context.Context, domain string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetTemplate(ctx context.Context, domain, name string) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) UpdateTemplate(ctx context.Context, domain, name string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

// Bounce suppression operations
func (m *MockWebhookClient) CreateBounce(ctx context.Context, domain string, bounce *bouncetypes.BounceParameters) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetBounce(ctx context.Context, domain, address string) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Complaint suppression operations
func (m *MockWebhookClient) CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetComplaint(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Unsubscribe suppression operations
func (m *MockWebhookClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWebhookClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func TestWebhookObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"WebhookExists": {
			reason: "Should return ResourceExists when webhook exists",
			args: args{
				mg: &v1beta1.Webhook{
					Spec: v1beta1.WebhookSpec{
						ForProvider: v1beta1.WebhookParameters{
							DomainRef: xpv1.Reference{
								Name: "example.com",
							},
							EventType: "delivered",
							URL:       "https://example.com/webhook",
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"WebhookNotFound": {
			reason: "Should return ResourceExists false when webhook not found",
			args: args{
				mg: &v1beta1.Webhook{
					Spec: v1beta1.WebhookSpec{
						ForProvider: v1beta1.WebhookParameters{
							DomainRef: xpv1.Reference{
								Name: "notfound.com",
							},
							EventType: "opened",
							URL:       "https://notfound.com/webhook",
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Setup mock client
			mockClient := &MockWebhookClient{}

			// Pre-populate with test webhook for "exists" test
			if name == "WebhookExists" {
				mockClient.webhooks = map[string]*v1beta1.WebhookObservation{
					"example.com/delivered": {
						ID:        "webhook_123",
						EventType: "delivered",
						URL:       "https://example.com/webhook",
						CreatedAt: "2025-01-01T00:00:00Z",
					},
				}
			}

			e := &external{service: mockClient}
			got, err := e.Observe(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.o.ResourceExists, got.ResourceExists)
				assert.Equal(t, tc.want.o.ResourceUpToDate, got.ResourceUpToDate)
			}
		})
	}
}

func TestWebhookCreate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulCreate": {
			reason: "Should successfully create webhook",
			args: args{
				mg: &v1beta1.Webhook{
					Spec: v1beta1.WebhookSpec{
						ForProvider: v1beta1.WebhookParameters{
							DomainRef: xpv1.Reference{
								Name: "new.com",
							},
							EventType: "clicked",
							URL:       "https://new.com/webhook",
							Username:  stringPtr("webhook_user"),
							Password:  stringPtr("webhook_pass"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockWebhookClient{}
			e := &external{service: mockClient}

			got, err := e.Create(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.o, got)
			}
		})
	}
}

func TestWebhookUpdate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulUpdate": {
			reason: "Should successfully update webhook",
			args: args{
				mg: &v1beta1.Webhook{
					Spec: v1beta1.WebhookSpec{
						ForProvider: v1beta1.WebhookParameters{
							DomainRef: xpv1.Reference{
								Name: "existing.com",
							},
							EventType: "delivered",
							URL:       "https://updated.com/webhook",
							Username:  stringPtr("updated_user"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalUpdate{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockWebhookClient{
				webhooks: map[string]*v1beta1.WebhookObservation{
					"existing.com/delivered": {
						ID:        "webhook_existing",
						EventType: "delivered",
						URL:       "https://existing.com/webhook",
						Username:  "original_user",
						CreatedAt: "2025-01-01T00:00:00Z",
					},
				},
			}
			e := &external{service: mockClient}

			got, err := e.Update(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.o, got)
			}
		})
	}
}

func TestWebhookDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulDelete": {
			reason: "Should successfully delete webhook",
			args: args{
				mg: &v1beta1.Webhook{
					Spec: v1beta1.WebhookSpec{
						ForProvider: v1beta1.WebhookParameters{
							DomainRef: xpv1.Reference{
								Name: "delete.com",
							},
							EventType: "opened",
							URL:       "https://delete.com/webhook",
						},
					},
				},
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockWebhookClient{
				webhooks: map[string]*v1beta1.WebhookObservation{
					"delete.com/opened": {
						ID:        "webhook_delete",
						EventType: "opened",
						URL:       "https://delete.com/webhook",
					},
				},
			}
			e := &external{service: mockClient}

			_, err := e.Delete(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				// Verify webhook was deleted
				_, exists := mockClient.webhooks["delete.com/opened"]
				assert.False(t, exists)
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
