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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	bouncetypes "github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	mailinglisttypes "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	webhooktypes "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
)

// MockDomainClient for testing
type MockDomainClient struct {
	domains map[string]*v1beta1.DomainObservation
	err     error
}

func (m *MockDomainClient) CreateDomain(ctx context.Context, domain *v1beta1.DomainParameters) (*v1beta1.DomainObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := &v1beta1.DomainObservation{
		ID:           domain.Name,
		State:        "active",
		CreatedAt:    "2025-01-01T00:00:00Z",
		SMTPLogin:    "postmaster@" + domain.Name,
		SMTPPassword: "generated-password",
		RequiredDNSRecords: []v1beta1.DNSRecord{
			{
				Name:     domain.Name,
				Type:     "TXT",
				Value:    "v=spf1 include:mailgun.org ~all",
				Priority: nil,
				Valid:    boolPtr(false),
			},
		},
	}

	if m.domains == nil {
		m.domains = make(map[string]*v1beta1.DomainObservation)
	}
	m.domains[domain.Name] = result

	return result, nil
}

func (m *MockDomainClient) GetDomain(ctx context.Context, name string) (*v1beta1.DomainObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	if domain, exists := m.domains[name]; exists {
		return domain, nil
	}

	return nil, errors.New("domain not found (404)")
}

func (m *MockDomainClient) UpdateDomain(ctx context.Context, name string, domain *v1beta1.DomainParameters) (*v1beta1.DomainObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	if existing, exists := m.domains[name]; exists {
		// Return the existing domain (no actual updates in mock)
		return existing, nil
	}

	return nil, errors.New("domain not found (404)")
}

func (m *MockDomainClient) DeleteDomain(ctx context.Context, name string) error {
	if m.err != nil {
		return m.err
	}

	delete(m.domains, name)
	return nil
}

// Implement other required client methods as no-ops
func (m *MockDomainClient) CreateMailingList(ctx context.Context, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetMailingList(ctx context.Context, address string) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) UpdateMailingList(ctx context.Context, address string, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

func (m *MockDomainClient) CreateRoute(ctx context.Context, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetRoute(ctx context.Context, id string) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) UpdateRoute(ctx context.Context, id string, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *MockDomainClient) CreateWebhook(ctx context.Context, domain string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetWebhook(ctx context.Context, domain, eventType string) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

func (m *MockDomainClient) CreateSMTPCredential(ctx context.Context, domain string, credential *smtpcredentialtypes.SMTPCredentialParameters) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetSMTPCredential(ctx context.Context, domain, login string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

func (m *MockDomainClient) CreateTemplate(ctx context.Context, domain string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetTemplate(ctx context.Context, domain, name string) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) UpdateTemplate(ctx context.Context, domain, name string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

// Bounce suppression operations
func (m *MockDomainClient) CreateBounce(ctx context.Context, domain string, bounce *bouncetypes.BounceParameters) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetBounce(ctx context.Context, domain, address string) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Complaint suppression operations
func (m *MockDomainClient) CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetComplaint(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Unsubscribe suppression operations
func (m *MockDomainClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDomainClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func TestDomainObserve(t *testing.T) {
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
		"DomainExists": {
			reason: "Should return ResourceExists when domain exists",
			args: args{
				mg: &v1beta1.Domain{
					Spec: v1beta1.DomainSpec{
						ForProvider: v1beta1.DomainParameters{
							Name: "example.com",
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: managed.ConnectionDetails{
						"smtp_login":    []byte("postmaster@example.com"),
						"smtp_password": []byte("generated-password"),
					},
				},
			},
		},
		"DomainNotFound": {
			reason: "Should return ResourceExists false when domain not found",
			args: args{
				mg: &v1beta1.Domain{
					Spec: v1beta1.DomainSpec{
						ForProvider: v1beta1.DomainParameters{
							Name: "notfound.com",
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
			mockClient := &MockDomainClient{}

			// Pre-populate with test domain for "exists" test
			if name == "DomainExists" {
				mockClient.domains = map[string]*v1beta1.DomainObservation{
					"example.com": {
						ID:           "example.com",
						State:        "active",
						CreatedAt:    "2025-01-01T00:00:00Z",
						SMTPLogin:    "postmaster@example.com",
						SMTPPassword: "generated-password",
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

				if tc.want.o.ConnectionDetails != nil {
					assert.Equal(t, tc.want.o.ConnectionDetails, got.ConnectionDetails)
				}
			}
		})
	}
}

func TestDomainCreate(t *testing.T) {
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
			reason: "Should successfully create domain",
			args: args{
				mg: &v1beta1.Domain{
					Spec: v1beta1.DomainSpec{
						ForProvider: v1beta1.DomainParameters{
							Name: "new.com",
							Type: stringPtr("sending"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						"smtp_login":    []byte("postmaster@new.com"),
						"smtp_password": []byte("generated-password"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockDomainClient{}
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

func TestDomainUpdate(t *testing.T) {
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
			reason: "Should successfully update domain",
			args: args{
				mg: &v1beta1.Domain{
					Spec: v1beta1.DomainSpec{
						ForProvider: v1beta1.DomainParameters{
							Name:       "existing.com",
							SpamAction: stringPtr("block"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalUpdate{
					ConnectionDetails: managed.ConnectionDetails{
						"smtp_login":    []byte("postmaster@existing.com"),
						"smtp_password": []byte("generated-password"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockDomainClient{
				domains: map[string]*v1beta1.DomainObservation{
					"existing.com": {
						ID:           "existing.com",
						State:        "active",
						CreatedAt:    "2025-01-01T00:00:00Z",
						SMTPLogin:    "postmaster@existing.com",
						SMTPPassword: "generated-password",
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

func TestDomainDelete(t *testing.T) {
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
			reason: "Should successfully delete domain",
			args: args{
				mg: &v1beta1.Domain{
					Spec: v1beta1.DomainSpec{
						ForProvider: v1beta1.DomainParameters{
							Name: "delete.com",
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
			mockClient := &MockDomainClient{
				domains: map[string]*v1beta1.DomainObservation{
					"delete.com": {
						ID:    "delete.com",
						State: "active",
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
				// Verify domain was deleted
				_, exists := mockClient.domains["delete.com"]
				assert.False(t, exists)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
