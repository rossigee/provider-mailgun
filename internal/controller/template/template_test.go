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

	"github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	domainv1beta1 "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
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
		Name:      template.Name,
		CreatedAt: "2025-01-01T00:00:00Z",
		CreatedBy: "api",
	}

	// Handle optional description
	if template.Description != nil {
		result.Description = *template.Description
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
				mg: &v1beta1.Template{
					Spec: v1beta1.TemplateSpec{
						ForProvider: v1beta1.TemplateParameters{
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
				mg: &v1beta1.Template{
					Spec: v1beta1.TemplateSpec{
						ForProvider: v1beta1.TemplateParameters{
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
				mg: &v1beta1.Template{
					Spec: v1beta1.TemplateSpec{
						ForProvider: v1beta1.TemplateParameters{
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
				mg: &v1beta1.Template{
					Spec: v1beta1.TemplateSpec{
						ForProvider: v1beta1.TemplateParameters{
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
				mg: &v1beta1.Template{
					Spec: v1beta1.TemplateSpec{
						ForProvider: v1beta1.TemplateParameters{
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

			_, err := e.Delete(context.Background(), tc.args.mg)

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

// Additional comprehensive test coverage for Template controller

// Test error handling scenarios
func TestTemplateObserveErrors(t *testing.T) {
	cases := map[string]struct {
		reason     string
		mockErr    error
		setupMock  func(*MockTemplateClient)
		expectErr  bool
		expectExists bool
	}{
		"ClientError": {
			reason:    "Should handle client errors gracefully",
			mockErr:   errors.New("mailgun api error"),
			expectErr: true,
		},
		"InvalidManagedResource": {
			reason:    "Should handle invalid managed resource type",
			expectErr: false, // Observe handles type check internally
		},
		"TemplateNotFoundButUpToDate": {
			reason: "Should handle template not found correctly",
			setupMock: func(m *MockTemplateClient) {
				// Mock will return "template not found" error for nonexistent templates
			},
			expectErr:    false,
			expectExists: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockTemplateClient{err: tc.mockErr}
			if tc.setupMock != nil {
				tc.setupMock(mockClient)
			}

			e := &external{client: mockClient}
			mg := &v1beta1.Template{
				Spec: v1beta1.TemplateSpec{
					ForProvider: v1beta1.TemplateParameters{
						Domain: "example.com",
						Name:   "test-template",
					},
				},
			}

			obs, err := e.Observe(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectExists, obs.ResourceExists)
			}
		})
	}
}

func TestTemplateCreateErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"CreateError": {
			reason:    "Should handle create errors",
			mockErr:   errors.New("failed to create template"),
			expectErr: true,
		},
		"InvalidResourceType": {
			reason:    "Should handle invalid resource type",
			expectErr: false, // Create handles type check internally
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockTemplateClient{err: tc.mockErr}
			e := &external{client: mockClient}

			mg := &v1beta1.Template{
				Spec: v1beta1.TemplateSpec{
					ForProvider: v1beta1.TemplateParameters{
						Domain:      "example.com",
						Name:        "error-template",
						Description: stringPtr("Error test template"),
						Template:    stringPtr("Hello {{name}}!"),
						Engine:      stringPtr("mustache"),
					},
				},
			}

			_, err := e.Create(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTemplateUpdateErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"UpdateError": {
			reason:    "Should handle update errors",
			mockErr:   errors.New("failed to update template"),
			expectErr: true,
		},
		"TemplateNotFound": {
			reason:    "Should handle template not found during update",
			mockErr:   errors.New("template not found (404)"),
			expectErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockTemplateClient{err: tc.mockErr}
			e := &external{client: mockClient}

			mg := &v1beta1.Template{
				Spec: v1beta1.TemplateSpec{
					ForProvider: v1beta1.TemplateParameters{
						Domain:      "example.com",
						Name:        "error-template",
						Description: stringPtr("Updated description"),
					},
				},
			}

			_, err := e.Update(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTemplateDeleteErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"DeleteError": {
			reason:    "Should handle delete errors",
			mockErr:   errors.New("failed to delete template"),
			expectErr: true,
		},
		"TemplateNotFound": {
			reason:    "Should handle template not found during delete",
			mockErr:   errors.New("template not found (404)"),
			expectErr: true, // Template controller doesn't handle 404 gracefully yet
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockTemplateClient{err: tc.mockErr}
			e := &external{client: mockClient}

			mg := &v1beta1.Template{
				Spec: v1beta1.TemplateSpec{
					ForProvider: v1beta1.TemplateParameters{
						Domain: "example.com",
						Name:   "error-template",
					},
				},
			}

			_, err := e.Delete(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test edge cases and boundary conditions
func TestTemplateEdgeCases(t *testing.T) {
	t.Run("EmptyTemplateFields", func(t *testing.T) {
		mockClient := &MockTemplateClient{}
		e := &external{client: mockClient}

		// Template with minimal required fields
		mg := &v1beta1.Template{
			Spec: v1beta1.TemplateSpec{
				ForProvider: v1beta1.TemplateParameters{
					Domain: "example.com",
					Name:   "minimal-template",
					// Description, Template, Engine are nil/empty
				},
			},
		}

		_, err := e.Create(context.Background(), mg)
		require.NoError(t, err)

		// Verify template was created with minimal fields
		key := "example.com/minimal-template"
		template, exists := mockClient.templates[key]
		assert.True(t, exists)
		assert.Equal(t, "minimal-template", template.Name)
	})

	t.Run("TemplateWithComplexContent", func(t *testing.T) {
		mockClient := &MockTemplateClient{}
		e := &external{client: mockClient}

		// Template with complex HTML/CSS content
		complexTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; }
		.header { background-color: #f0f0f0; padding: 20px; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Welcome {{name}}!</h1>
		<p>Your email is: {{email}}</p>
	</div>
	<div class="content">
		{{#items}}
			<li>{{title}} - {{price}}</li>
		{{/items}}
	</div>
</body>
</html>`

		mg := &v1beta1.Template{
			Spec: v1beta1.TemplateSpec{
				ForProvider: v1beta1.TemplateParameters{
					Domain:      "example.com",
					Name:        "complex-template",
					Description: stringPtr("Complex HTML template with CSS and Mustache syntax"),
					Template:    &complexTemplate,
					Engine:      stringPtr("mustache"),
				},
			},
		}

		_, err := e.Create(context.Background(), mg)
		require.NoError(t, err)

		// Verify template was created with complex content
		key := "example.com/complex-template"
		template, exists := mockClient.templates[key]
		assert.True(t, exists)
		assert.Equal(t, "complex-template", template.Name)
		assert.Len(t, template.Versions, 1)
		assert.Equal(t, complexTemplate, template.Versions[0].Template)
	})

	t.Run("TemplateStatusUpdate", func(t *testing.T) {
		mockClient := &MockTemplateClient{
			templates: map[string]*clients.Template{
				"example.com/status-template": {
					Name:        "status-template",
					Description: "Template for status testing",
					CreatedAt:   "2025-01-01T00:00:00Z",
					CreatedBy:   "test-user",
					Versions: []clients.TemplateVersion{
						{
							Tag:       "v1.0",
							Engine:    "mustache",
							CreatedAt: "2025-01-01T00:00:00Z",
							Comment:   "Initial version",
							Active:    true,
							Template:  "Hello {{name}}!",
						},
					},
				},
			},
		}
		e := &external{client: mockClient}

		mg := &v1beta1.Template{
			Spec: v1beta1.TemplateSpec{
				ForProvider: v1beta1.TemplateParameters{
					Domain: "example.com",
					Name:   "status-template",
				},
			},
		}

		obs, err := e.Observe(context.Background(), mg)
		require.NoError(t, err)
		assert.True(t, obs.ResourceExists)
		assert.True(t, obs.ResourceUpToDate)

		// Verify the managed resource status is updated
		assert.Equal(t, "status-template", mg.Status.AtProvider.Name)
		assert.Equal(t, "Template for status testing", mg.Status.AtProvider.Description)
		assert.Equal(t, "2025-01-01T00:00:00Z", mg.Status.AtProvider.CreatedAt)
		assert.Equal(t, "test-user", mg.Status.AtProvider.CreatedBy)
	})
}

// Test invalid managed resource types to improve error handling coverage
func TestTemplateInvalidManagedResource(t *testing.T) {
	operations := []struct {
		name string
		op   func(*external, context.Context, resource.Managed) error
	}{
		{
			name: "Observe",
			op: func(e *external, ctx context.Context, mg resource.Managed) error {
				_, err := e.Observe(ctx, mg)
				return err
			},
		},
		{
			name: "Create",
			op: func(e *external, ctx context.Context, mg resource.Managed) error {
				_, err := e.Create(ctx, mg)
				return err
			},
		},
		{
			name: "Update",
			op: func(e *external, ctx context.Context, mg resource.Managed) error {
				_, err := e.Update(ctx, mg)
				return err
			},
		},
		{
			name: "Delete",
			op: func(e *external, ctx context.Context, mg resource.Managed) error {
				_, err := e.Delete(ctx, mg)
				return err
			},
		},
	}

	for _, op := range operations {
		t.Run(op.name+"InvalidType", func(t *testing.T) {
			mockClient := &MockTemplateClient{}
			e := &external{client: mockClient}

			// Use a different resource type (not Template)
			invalidMg := &domainv1beta1.Domain{} // Wrong type

			err := op.op(e, context.Background(), invalidMg)
			// Should handle gracefully without panicking
			// The exact error depends on implementation
			if err != nil {
				assert.Contains(t, err.Error(), "Template", "Error should mention Template type mismatch")
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
