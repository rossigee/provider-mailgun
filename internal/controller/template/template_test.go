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

package template

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-mailgun/apis/template/v1alpha1"
	"github.com/crossplane-contrib/provider-mailgun/internal/clients"
)

// MockTemplateClient for testing
type MockTemplateClient struct {
	templates map[string]*clients.Template
	err       error
}

func (m *MockTemplateClient) CreateTemplate(ctx context.Context, domain string, template *clients.TemplateSpec) (*clients.Template, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + template.Name
	result := &clients.Template{
		Name:        template.Name,
		Description: *template.Description,
		CreatedAt:   "2025-01-01T00:00:00Z",
		CreatedBy:   "api",
	}

	if template.Template != nil {
		result.Versions = []clients.TemplateVersion{
			{
				Tag:       "v1.0",
				Engine:    "mustache",
				CreatedAt: "2025-01-01T00:00:00Z",
				Comment:   "Initial version",
				Active:    true,
				Template:  *template.Template,
			},
		}
	}

	if m.templates == nil {
		m.templates = make(map[string]*clients.Template)
	}
	m.templates[key] = result

	return result, nil
}

func (m *MockTemplateClient) GetTemplate(ctx context.Context, domain, name string) (*clients.Template, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + name
	if template, exists := m.templates[key]; exists {
		return template, nil
	}

	return nil, errors.New("template not found (404)")
}

func (m *MockTemplateClient) UpdateTemplate(ctx context.Context, domain, name string, template *clients.TemplateSpec) (*clients.Template, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + name
	if existing, exists := m.templates[key]; exists {
		if template.Description != nil {
			existing.Description = *template.Description
		}
		return existing, nil
	}

	return nil, errors.New("template not found (404)")
}

func (m *MockTemplateClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	if m.err != nil {
		return m.err
	}

	key := domain + "/" + name
	delete(m.templates, key)
	return nil
}

// Implement other required client methods as no-ops
func (m *MockTemplateClient) CreateDomain(ctx context.Context, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetDomain(ctx context.Context, name string) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) UpdateDomain(ctx context.Context, name string, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockTemplateClient) CreateMailingList(ctx context.Context, list *clients.MailingListSpec) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetMailingList(ctx context.Context, address string) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) UpdateMailingList(ctx context.Context, address string, list *clients.MailingListSpec) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

func (m *MockTemplateClient) CreateRoute(ctx context.Context, route *clients.RouteSpec) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetRoute(ctx context.Context, id string) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) UpdateRoute(ctx context.Context, id string, route *clients.RouteSpec) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *MockTemplateClient) CreateWebhook(ctx context.Context, domain string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetWebhook(ctx context.Context, domain, eventType string) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

func (m *MockTemplateClient) CreateSMTPCredential(ctx context.Context, domain string, credential *clients.SMTPCredentialSpec) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetSMTPCredential(ctx context.Context, domain, login string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

// Bounce suppression operations
func (m *MockTemplateClient) CreateBounce(ctx context.Context, domain string, bounce *clients.BounceSpec) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetBounce(ctx context.Context, domain, address string) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Complaint suppression operations
func (m *MockTemplateClient) CreateComplaint(ctx context.Context, domain string, complaint *clients.ComplaintSpec) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetComplaint(ctx context.Context, domain, address string) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Unsubscribe suppression operations
func (m *MockTemplateClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe *clients.UnsubscribeSpec) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) GetUnsubscribe(ctx context.Context, domain, address string) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockTemplateClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func TestTemplateObserve(t *testing.T) {
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
		"TemplateExists": {
			reason: "Should return ResourceExists when template exists",
			args: args{
				mg: &v1alpha1.Template{
					Spec: v1alpha1.TemplateSpec{
						ForProvider: v1alpha1.TemplateParameters{
							Domain: "example.com",
							Name:   "test-template",
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
		"TemplateNotFound": {
			reason: "Should return ResourceExists false when template not found",
			args: args{
				mg: &v1alpha1.Template{
					Spec: v1alpha1.TemplateSpec{
						ForProvider: v1alpha1.TemplateParameters{
							Domain: "example.com",
							Name:   "nonexistent-template",
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
			mockClient := &MockTemplateClient{}

			// Pre-populate with test template for "exists" test
			if name == "TemplateExists" {
				mockClient.templates = map[string]*clients.Template{
					"example.com/test-template": {
						Name:        "test-template",
						Description: "Test template",
						CreatedAt:   "2025-01-01T00:00:00Z",
						CreatedBy:   "api",
						Versions: []clients.TemplateVersion{
							{
								Tag:       "v1.0",
								Engine:    "mustache",
								CreatedAt: "2025-01-01T00:00:00Z",
								Comment:   "Initial version",
								Active:    true,
							},
						},
					},
				}
			}

			e := &external{client: mockClient}
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

func TestTemplateCreate(t *testing.T) {
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
			reason: "Should successfully create template",
			args: args{
				mg: &v1alpha1.Template{
					Spec: v1alpha1.TemplateSpec{
						ForProvider: v1alpha1.TemplateParameters{
							Domain:      "example.com",
							Name:        "new-template",
							Description: stringPtr("New test template"),
							Template:    stringPtr("<h1>Hello {{name}}!</h1>"),
							Engine:      stringPtr("mustache"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockTemplateClient{}
			e := &external{client: mockClient}

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

func TestTemplateUpdate(t *testing.T) {
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
			reason: "Should successfully update template description",
			args: args{
				mg: &v1alpha1.Template{
					Spec: v1alpha1.TemplateSpec{
						ForProvider: v1alpha1.TemplateParameters{
							Domain:      "example.com",
							Name:        "existing-template",
							Description: stringPtr("Updated description"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockTemplateClient{
				templates: map[string]*clients.Template{
					"example.com/existing-template": {
						Name:        "existing-template",
						Description: "Old description",
						CreatedAt:   "2025-01-01T00:00:00Z",
						CreatedBy:   "api",
					},
				},
			}
			e := &external{client: mockClient}

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

func TestTemplateDelete(t *testing.T) {
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
			reason: "Should successfully delete template",
			args: args{
				mg: &v1alpha1.Template{
					Spec: v1alpha1.TemplateSpec{
						ForProvider: v1alpha1.TemplateParameters{
							Domain: "example.com",
							Name:   "delete-template",
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
			mockClient := &MockTemplateClient{
				templates: map[string]*clients.Template{
					"example.com/delete-template": {
						Name:        "delete-template",
						Description: "Template to delete",
						CreatedAt:   "2025-01-01T00:00:00Z",
						CreatedBy:   "api",
					},
				},
			}
			e := &external{client: mockClient}

			err := e.Delete(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				// Verify template was deleted
				_, exists := mockClient.templates["example.com/delete-template"]
				assert.False(t, exists)
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
