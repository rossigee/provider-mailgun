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

package controller

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglistv1beta1 "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routev1beta1 "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpv1beta1 "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatev1beta1 "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
	"github.com/rossigee/provider-mailgun/internal/controller/domain"
	"github.com/rossigee/provider-mailgun/internal/controller/mailinglist"
	"github.com/rossigee/provider-mailgun/internal/controller/route"
	"github.com/rossigee/provider-mailgun/internal/controller/smtpcredential"
	"github.com/rossigee/provider-mailgun/internal/controller/template"
)

// IntegrationMockClient provides a unified mock for all Mailgun resources
// Used for testing multi-resource workflows and dependencies
type IntegrationMockClient struct {
	domains         map[string]*clients.Domain
	mailingLists    map[string]*clients.MailingList
	routes          map[string]*clients.Route
	webhooks        map[string]*clients.Webhook
	smtpCredentials map[string]*clients.SMTPCredential
	templates       map[string]*clients.Template
	err             error
}

func NewIntegrationMockClient() *IntegrationMockClient {
	return &IntegrationMockClient{
		domains:         make(map[string]*clients.Domain),
		mailingLists:    make(map[string]*clients.MailingList),
		routes:          make(map[string]*clients.Route),
		webhooks:        make(map[string]*clients.Webhook),
		smtpCredentials: make(map[string]*clients.SMTPCredential),
		templates:       make(map[string]*clients.Template),
	}
}

// Domain operations
func (m *IntegrationMockClient) CreateDomain(ctx context.Context, domain *clients.DomainSpec) (*clients.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := &clients.Domain{
		Name:              domain.Name,
		State:             "active",
		CreatedAt:         "2025-01-01T00:00:00Z",
		SMTPLogin:         "postmaster@" + domain.Name,
		RequiredDNSRecords: []clients.DNSRecord{
			{
				Type:  "TXT",
				Name:  domain.Name,
				Value: "v=spf1 include:mailgun.org ~all",
			},
		},
	}

	m.domains[domain.Name] = result
	return result, nil
}

func (m *IntegrationMockClient) GetDomain(ctx context.Context, name string) (*clients.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}

	if domain, exists := m.domains[name]; exists {
		return domain, nil
	}
	return nil, errors.New("domain not found (404)")
}

func (m *IntegrationMockClient) UpdateDomain(ctx context.Context, name string, domain *clients.DomainSpec) (*clients.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}

	if existing, exists := m.domains[name]; exists {
		// Domain updates would modify state, but for simplicity we just return existing
		return existing, nil
	}
	return nil, errors.New("domain not found (404)")
}

func (m *IntegrationMockClient) DeleteDomain(ctx context.Context, name string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.domains, name)
	return nil
}

// MailingList operations
func (m *IntegrationMockClient) CreateMailingList(ctx context.Context, list *clients.MailingListSpec) (*clients.MailingList, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := &clients.MailingList{
		Address:         list.Address,
		Name:            "Test List",
		Description:     "Test mailing list",
		AccessLevel:     "readonly",
		ReplyPreference: "list",
		CreatedAt:       "2025-01-01T00:00:00Z",
		MembersCount:    0,
	}

	if list.Name != nil {
		result.Name = *list.Name
	}
	if list.Description != nil {
		result.Description = *list.Description
	}
	if list.AccessLevel != nil {
		result.AccessLevel = *list.AccessLevel
	}
	if list.ReplyPreference != nil {
		result.ReplyPreference = *list.ReplyPreference
	}

	m.mailingLists[list.Address] = result
	return result, nil
}

func (m *IntegrationMockClient) GetMailingList(ctx context.Context, address string) (*clients.MailingList, error) {
	if m.err != nil {
		return nil, m.err
	}

	if list, exists := m.mailingLists[address]; exists {
		return list, nil
	}
	return nil, errors.New("mailing list not found (404)")
}

func (m *IntegrationMockClient) UpdateMailingList(ctx context.Context, address string, list *clients.MailingListSpec) (*clients.MailingList, error) {
	if m.err != nil {
		return nil, m.err
	}

	if existing, exists := m.mailingLists[address]; exists {
		if list.Name != nil {
			existing.Name = *list.Name
		}
		if list.Description != nil {
			existing.Description = *list.Description
		}
		if list.AccessLevel != nil {
			existing.AccessLevel = *list.AccessLevel
		}
		if list.ReplyPreference != nil {
			existing.ReplyPreference = *list.ReplyPreference
		}
		return existing, nil
	}
	return nil, errors.New("mailing list not found (404)")
}

func (m *IntegrationMockClient) DeleteMailingList(ctx context.Context, address string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.mailingLists, address)
	return nil
}

// Route operations
func (m *IntegrationMockClient) CreateRoute(ctx context.Context, route *clients.RouteSpec) (*clients.Route, error) {
	if m.err != nil {
		return nil, m.err
	}

	id := "route_" + generateRandomID()
	result := &clients.Route{
		ID:          id,
		Priority:    0,
		Description: "Test route",
		Expression:  route.Expression,
		CreatedAt:   "2025-01-01T00:00:00Z",
	}

	if route.Priority != nil {
		result.Priority = *route.Priority
	}
	if route.Description != nil {
		result.Description = *route.Description
	}
	if len(route.Actions) > 0 {
		result.Actions = make([]clients.RouteAction, len(route.Actions))
		for i, action := range route.Actions {
			result.Actions[i] = clients.RouteAction{
				Type:        action.Type,
				Destination: action.Destination,
			}
		}
	}

	m.routes[id] = result
	return result, nil
}

func (m *IntegrationMockClient) GetRoute(ctx context.Context, id string) (*clients.Route, error) {
	if m.err != nil {
		return nil, m.err
	}

	if route, exists := m.routes[id]; exists {
		return route, nil
	}
	return nil, errors.New("route not found (404)")
}

func (m *IntegrationMockClient) UpdateRoute(ctx context.Context, id string, route *clients.RouteSpec) (*clients.Route, error) {
	if m.err != nil {
		return nil, m.err
	}

	if existing, exists := m.routes[id]; exists {
		if route.Priority != nil {
			existing.Priority = *route.Priority
		}
		if route.Description != nil {
			existing.Description = *route.Description
		}
		existing.Expression = route.Expression
		if len(route.Actions) > 0 {
			existing.Actions = make([]clients.RouteAction, len(route.Actions))
			for i, action := range route.Actions {
				existing.Actions[i] = clients.RouteAction{
					Type:        action.Type,
					Destination: action.Destination,
				}
			}
		}
		return existing, nil
	}
	return nil, errors.New("route not found (404)")
}

func (m *IntegrationMockClient) DeleteRoute(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.routes, id)
	return nil
}

// Webhook operations
func (m *IntegrationMockClient) CreateWebhook(ctx context.Context, domain string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Check if domain exists
	if _, exists := m.domains[domain]; !exists {
		return nil, errors.New("domain not found (404)")
	}

	key := domain + "/" + webhook.EventType
	result := &clients.Webhook{
		EventType: webhook.EventType,
		URL:       webhook.URL,
		CreatedAt: "2025-01-01T00:00:00Z",
	}

	m.webhooks[key] = result
	return result, nil
}

func (m *IntegrationMockClient) GetWebhook(ctx context.Context, domain, eventType string) (*clients.Webhook, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + eventType
	if webhook, exists := m.webhooks[key]; exists {
		return webhook, nil
	}
	return nil, errors.New("webhook not found (404)")
}

func (m *IntegrationMockClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + eventType
	if existing, exists := m.webhooks[key]; exists {
		existing.URL = webhook.URL
		return existing, nil
	}
	return nil, errors.New("webhook not found (404)")
}

func (m *IntegrationMockClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	if m.err != nil {
		return m.err
	}
	key := domain + "/" + eventType
	delete(m.webhooks, key)
	return nil
}

// SMTP Credential operations
func (m *IntegrationMockClient) CreateSMTPCredential(ctx context.Context, domain string, credential *clients.SMTPCredentialSpec) (*clients.SMTPCredential, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Check if domain exists
	if _, exists := m.domains[domain]; !exists {
		return nil, errors.New("domain not found (404)")
	}

	key := domain + "/" + credential.Login
	result := &clients.SMTPCredential{
		Login:     credential.Login,
		Password:  "generated_password_123",
		CreatedAt: "2025-01-01T00:00:00Z",
		State:     "active",
	}

	m.smtpCredentials[key] = result
	return result, nil
}

func (m *IntegrationMockClient) GetSMTPCredential(ctx context.Context, domain, login string) (*clients.SMTPCredential, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + login
	if credential, exists := m.smtpCredentials[key]; exists {
		return credential, nil
	}
	return nil, errors.New("SMTP credential not found (404)")
}

func (m *IntegrationMockClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*clients.SMTPCredential, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + login
	if existing, exists := m.smtpCredentials[key]; exists {
		existing.Password = password
		return existing, nil
	}
	return nil, errors.New("SMTP credential not found (404)")
}

func (m *IntegrationMockClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	if m.err != nil {
		return m.err
	}
	key := domain + "/" + login
	delete(m.smtpCredentials, key)
	return nil
}

// Template operations
func (m *IntegrationMockClient) CreateTemplate(ctx context.Context, domain string, template *clients.TemplateSpec) (*clients.Template, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Check if domain exists
	if _, exists := m.domains[domain]; !exists {
		return nil, errors.New("domain not found (404)")
	}

	key := domain + "/" + template.Name
	result := &clients.Template{
		Name:      template.Name,
		CreatedAt: "2025-01-01T00:00:00Z",
		CreatedBy: "api",
	}

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

	m.templates[key] = result
	return result, nil
}

func (m *IntegrationMockClient) GetTemplate(ctx context.Context, domain, name string) (*clients.Template, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + name
	if template, exists := m.templates[key]; exists {
		return template, nil
	}
	return nil, errors.New("template not found (404)")
}

func (m *IntegrationMockClient) UpdateTemplate(ctx context.Context, domain, name string, template *clients.TemplateSpec) (*clients.Template, error) {
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

func (m *IntegrationMockClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	if m.err != nil {
		return m.err
	}
	key := domain + "/" + name
	delete(m.templates, key)
	return nil
}

// Suppression operations (simplified for integration tests)
func (m *IntegrationMockClient) CreateBounce(ctx context.Context, domain string, bounce *clients.BounceSpec) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *IntegrationMockClient) GetBounce(ctx context.Context, domain, address string) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *IntegrationMockClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func (m *IntegrationMockClient) CreateComplaint(ctx context.Context, domain string, complaint *clients.ComplaintSpec) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *IntegrationMockClient) GetComplaint(ctx context.Context, domain, address string) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *IntegrationMockClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func (m *IntegrationMockClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe *clients.UnsubscribeSpec) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *IntegrationMockClient) GetUnsubscribe(ctx context.Context, domain, address string) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *IntegrationMockClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Helper functions
func generateRandomID() string {
	return "123456"
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

// Integration Test Scenarios

// TestCompleteEmailSetupWorkflow tests a complete email setup scenario:
// 1. Create a domain
// 2. Set up SMTP credentials for the domain
// 3. Create a mailing list for the domain
// 4. Set up email templates
// 5. Configure routing rules
// 6. Set up webhooks for notifications
func TestCompleteEmailSetupWorkflow(t *testing.T) {
	mockClient := NewIntegrationMockClient()
	ctx := context.Background()

	domainName := "integration-test.com"

	// Step 1: Create a domain
	t.Run("CreateDomain", func(t *testing.T) {
		domainExternal := &domain.ExternalForTesting{Client: mockClient}

		domainMg := &v1beta1.Domain{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-domain",
			},
			Spec: v1beta1.DomainSpec{
				ForProvider: v1beta1.DomainParameters{
					Name: domainName,
				},
			},
		}

		obs, err := domainExternal.Observe(ctx, domainMg)
		require.NoError(t, err)
		assert.False(t, obs.ResourceExists)

		creation, err := domainExternal.Create(ctx, domainMg)
		require.NoError(t, err)
		assert.NotNil(t, creation)

		// Verify domain was created - check annotation for external name
		assert.NotEmpty(t, domainMg.GetAnnotations())
	})

	// Step 2: Set up SMTP credentials for the domain
	t.Run("CreateSMTPCredentials", func(t *testing.T) {
		smtpExternal := &smtpcredential.ExternalForTesting{Client: mockClient}

		smtpMg := &smtpv1beta1.SMTPCredential{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-smtp",
			},
			Spec: smtpv1beta1.SMTPCredentialSpec{
				ForProvider: smtpv1beta1.SMTPCredentialParameters{
					Domain: domainName,
					Login:  "admin",
				},
			},
		}

		obs, err := smtpExternal.Observe(ctx, smtpMg)
		require.NoError(t, err)
		assert.False(t, obs.ResourceExists)

		creation, err := smtpExternal.Create(ctx, smtpMg)
		require.NoError(t, err)
		assert.NotNil(t, creation)

		// Verify SMTP credential was created
		assert.Equal(t, "admin", smtpMg.Status.AtProvider.Login)
	})

	// Step 3: Create a mailing list for the domain
	t.Run("CreateMailingList", func(t *testing.T) {
		mailingListExternal := &mailinglist.ExternalForTesting{Client: mockClient}

		mailingListMg := &mailinglistv1beta1.MailingList{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-list",
			},
			Spec: mailinglistv1beta1.MailingListSpec{
				ForProvider: mailinglistv1beta1.MailingListParameters{
					Address:         "newsletter@" + domainName,
					Name:            stringPtr("Newsletter List"),
					Description:     stringPtr("Main newsletter mailing list"),
					AccessLevel:     stringPtr("members"),
					ReplyPreference: stringPtr("list"),
				},
			},
		}

		obs, err := mailingListExternal.Observe(ctx, mailingListMg)
		require.NoError(t, err)
		assert.False(t, obs.ResourceExists)

		creation, err := mailingListExternal.Create(ctx, mailingListMg)
		require.NoError(t, err)
		assert.NotNil(t, creation)

		// Verify mailing list was created
		assert.Equal(t, "newsletter@"+domainName, mailingListMg.Status.AtProvider.Address)
		assert.Equal(t, "Newsletter List", mailingListMg.Status.AtProvider.Name)
	})

	// Step 4: Set up email templates
	t.Run("CreateEmailTemplate", func(t *testing.T) {
		templateExternal := &template.ExternalForTesting{Client: mockClient}

		templateMg := &templatev1beta1.Template{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-template",
			},
			Spec: templatev1beta1.TemplateSpec{
				ForProvider: templatev1beta1.TemplateParameters{
					Domain:      domainName,
					Name:        "welcome-email",
					Description: stringPtr("Welcome email template"),
					Template:    stringPtr("Hello {{name}}, welcome to our service!"),
					Engine:      stringPtr("mustache"),
				},
			},
		}

		obs, err := templateExternal.Observe(ctx, templateMg)
		require.NoError(t, err)
		assert.False(t, obs.ResourceExists)

		creation, err := templateExternal.Create(ctx, templateMg)
		require.NoError(t, err)
		assert.NotNil(t, creation)

		// Call Observe again to populate status after creation
		obs, err = templateExternal.Observe(ctx, templateMg)
		require.NoError(t, err)
		assert.True(t, obs.ResourceExists)

		// Verify template was created and status populated
		assert.Equal(t, "welcome-email", templateMg.Status.AtProvider.Name)
		assert.Equal(t, "Welcome email template", templateMg.Status.AtProvider.Description)
	})

	// Step 5: Configure routing rules
	t.Run("CreateRoutingRule", func(t *testing.T) {
		routeExternal := &route.ExternalForTesting{Client: mockClient}

		routeMg := &routev1beta1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-route",
			},
			Spec: routev1beta1.RouteSpec{
				ForProvider: routev1beta1.RouteParameters{
					Priority:    intPtr(10),
					Description: stringPtr("Newsletter routing rule"),
					Expression:  "match_recipient(\"newsletter@" + domainName + "\")",
					Actions: []routev1beta1.RouteAction{
						{
							Type:        "forward",
							Destination: stringPtr("admin@" + domainName),
						},
					},
				},
			},
		}

		obs, err := routeExternal.Observe(ctx, routeMg)
		require.NoError(t, err)
		assert.False(t, obs.ResourceExists)

		creation, err := routeExternal.Create(ctx, routeMg)
		require.NoError(t, err)
		assert.NotNil(t, creation)

		// Verify route was created
		assert.Equal(t, 10, routeMg.Status.AtProvider.Priority)
		assert.Equal(t, "Newsletter routing rule", routeMg.Status.AtProvider.Description)
		assert.Len(t, routeMg.Status.AtProvider.Actions, 1)
	})

	// Step 6: Set up webhooks for notifications
	t.Run("CreateWebhooks", func(t *testing.T) {
		// For integration tests, we'll test webhook creation directly through the mock client
		// since the actual webhook controller requires domain resolution which is complex to mock
		webhook, err := mockClient.CreateWebhook(ctx, domainName, &clients.WebhookSpec{
			EventType: "delivered",
			URL:       "https://api.example.com/webhooks/mailgun/delivered",
		})
		require.NoError(t, err)
		assert.Equal(t, "delivered", webhook.EventType)
		assert.Equal(t, "https://api.example.com/webhooks/mailgun/delivered", webhook.URL)
	})

	// Verify all resources exist and are properly configured
	t.Run("VerifyCompleteSetup", func(t *testing.T) {
		// Verify domain exists
		domain, err := mockClient.GetDomain(ctx, domainName)
		require.NoError(t, err)
		assert.Equal(t, domainName, domain.Name)
		assert.Equal(t, "active", domain.State)

		// Verify SMTP credential exists
		smtp, err := mockClient.GetSMTPCredential(ctx, domainName, "admin")
		require.NoError(t, err)
		assert.Equal(t, "admin", smtp.Login)

		// Verify mailing list exists
		mailingList, err := mockClient.GetMailingList(ctx, "newsletter@"+domainName)
		require.NoError(t, err)
		assert.Equal(t, "newsletter@"+domainName, mailingList.Address)

		// Verify template exists
		template, err := mockClient.GetTemplate(ctx, domainName, "welcome-email")
		require.NoError(t, err)
		assert.Equal(t, "welcome-email", template.Name)

		// Verify webhook exists
		webhook, err := mockClient.GetWebhook(ctx, domainName, "delivered")
		require.NoError(t, err)
		assert.Equal(t, "delivered", webhook.EventType)

		// Verify route exists (we need to check in the mock's routes map)
		assert.Len(t, mockClient.routes, 1)
		for _, route := range mockClient.routes {
			assert.Contains(t, route.Expression, domainName)
			assert.Len(t, route.Actions, 1)
			assert.Equal(t, "forward", route.Actions[0].Type)
		}
	})
}

// TestDomainDependencyWorkflow tests scenarios where resources depend on domains
func TestDomainDependencyWorkflow(t *testing.T) {
	mockClient := NewIntegrationMockClient()
	ctx := context.Background()

	domainName := "dependency-test.com"

	// Test creating dependent resources before domain exists (should fail)
	t.Run("CreateDependentResourcesWithoutDomain", func(t *testing.T) {
		// Try to create SMTP credential without domain
		smtpExternal := &smtpcredential.ExternalForTesting{Client: mockClient}
		smtpMg := &smtpv1beta1.SMTPCredential{
			Spec: smtpv1beta1.SMTPCredentialSpec{
				ForProvider: smtpv1beta1.SMTPCredentialParameters{
					Domain: domainName,
					Login:  "test",
				},
			},
		}

		_, err := smtpExternal.Create(ctx, smtpMg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "domain not found")

		// Try to create webhook without domain
		_, err = mockClient.CreateWebhook(ctx, domainName, &clients.WebhookSpec{
			EventType: "delivered",
			URL:       "https://example.com/webhook",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "domain not found")

		// Try to create template without domain
		templateExternal := &template.ExternalForTesting{Client: mockClient}
		templateMg := &templatev1beta1.Template{
			Spec: templatev1beta1.TemplateSpec{
				ForProvider: templatev1beta1.TemplateParameters{
					Domain: domainName,
					Name:   "test-template",
				},
			},
		}

		_, err = templateExternal.Create(ctx, templateMg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "domain not found")
	})

	// Create domain first
	t.Run("CreateDomain", func(t *testing.T) {
		domainExternal := &domain.ExternalForTesting{Client: mockClient}
		domainMg := &v1beta1.Domain{
			Spec: v1beta1.DomainSpec{
				ForProvider: v1beta1.DomainParameters{
					Name: domainName,
				},
			},
		}

		_, err := domainExternal.Create(ctx, domainMg)
		require.NoError(t, err)
	})

	// Now try creating dependent resources (should succeed)
	t.Run("CreateDependentResourcesWithDomain", func(t *testing.T) {
		// Create SMTP credential
		smtpExternal := &smtpcredential.ExternalForTesting{Client: mockClient}
		smtpMg := &smtpv1beta1.SMTPCredential{
			Spec: smtpv1beta1.SMTPCredentialSpec{
				ForProvider: smtpv1beta1.SMTPCredentialParameters{
					Domain: domainName,
					Login:  "test",
				},
			},
		}

		_, err := smtpExternal.Create(ctx, smtpMg)
		require.NoError(t, err)

		// Create webhook
		_, err = mockClient.CreateWebhook(ctx, domainName, &clients.WebhookSpec{
			EventType: "delivered",
			URL:       "https://example.com/webhook",
		})
		require.NoError(t, err)

		// Create template
		templateExternal := &template.ExternalForTesting{Client: mockClient}
		templateMg := &templatev1beta1.Template{
			Spec: templatev1beta1.TemplateSpec{
				ForProvider: templatev1beta1.TemplateParameters{
					Domain: domainName,
					Name:   "test-template",
				},
			},
		}

		_, err = templateExternal.Create(ctx, templateMg)
		require.NoError(t, err)
	})
}

// TestResourceUpdateCascade tests updating resources and their dependencies
func TestResourceUpdateCascade(t *testing.T) {
	mockClient := NewIntegrationMockClient()
	ctx := context.Background()

	domainName := "update-test.com"

	// Setup initial resources
	t.Run("SetupResources", func(t *testing.T) {
		// Create domain
		domainExternal := &domain.ExternalForTesting{Client: mockClient}
		domainMg := &v1beta1.Domain{
			Spec: v1beta1.DomainSpec{
				ForProvider: v1beta1.DomainParameters{
					Name: domainName,
				},
			},
		}
		_, err := domainExternal.Create(ctx, domainMg)
		require.NoError(t, err)

		// Create mailing list
		mailingListExternal := &mailinglist.ExternalForTesting{Client: mockClient}
		mailingListMg := &mailinglistv1beta1.MailingList{
			Spec: mailinglistv1beta1.MailingListSpec{
				ForProvider: mailinglistv1beta1.MailingListParameters{
					Address:     "test@" + domainName,
					Name:        stringPtr("Original Name"),
					AccessLevel: stringPtr("readonly"),
				},
			},
		}
		_, err = mailingListExternal.Create(ctx, mailingListMg)
		require.NoError(t, err)
	})

	// Test updating domain configuration
	t.Run("UpdateDomainConfiguration", func(t *testing.T) {
		domainExternal := &domain.ExternalForTesting{Client: mockClient}
		domainMg := &v1beta1.Domain{
			Spec: v1beta1.DomainSpec{
				ForProvider: v1beta1.DomainParameters{
					Name: domainName,
				},
			},
		}

		_, err := domainExternal.Update(ctx, domainMg)
		require.NoError(t, err)

		// Verify domain was updated
		domain, err := mockClient.GetDomain(ctx, domainName)
		require.NoError(t, err)
		assert.Equal(t, domainName, domain.Name)
	})

	// Test updating mailing list settings
	t.Run("UpdateMailingListSettings", func(t *testing.T) {
		mailingListExternal := &mailinglist.ExternalForTesting{Client: mockClient}
		mailingListMg := &mailinglistv1beta1.MailingList{
			Spec: mailinglistv1beta1.MailingListSpec{
				ForProvider: mailinglistv1beta1.MailingListParameters{
					Address:     "test@" + domainName,
					Name:        stringPtr("Updated Name"),
					AccessLevel: stringPtr("members"), // Change from readonly to members
				},
			},
		}

		_, err := mailingListExternal.Update(ctx, mailingListMg)
		require.NoError(t, err)

		// Verify mailing list was updated
		mailingList, err := mockClient.GetMailingList(ctx, "test@"+domainName)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", mailingList.Name)
		assert.Equal(t, "members", mailingList.AccessLevel)
	})
}

// TestResourceDeletionOrder tests proper deletion order to avoid dependency issues
func TestResourceDeletionOrder(t *testing.T) {
	mockClient := NewIntegrationMockClient()
	ctx := context.Background()

	domainName := "deletion-test.com"

	// Setup complete infrastructure
	t.Run("SetupCompleteInfrastructure", func(t *testing.T) {
		// Create domain
		domainExternal := &domain.ExternalForTesting{Client: mockClient}
		domainMg := &v1beta1.Domain{
			Spec: v1beta1.DomainSpec{
				ForProvider: v1beta1.DomainParameters{
					Name: domainName,
				},
			},
		}
		_, err := domainExternal.Create(ctx, domainMg)
		require.NoError(t, err)

		// Create dependent resources
		smtpExternal := &smtpcredential.ExternalForTesting{Client: mockClient}
		smtpMg := &smtpv1beta1.SMTPCredential{
			Spec: smtpv1beta1.SMTPCredentialSpec{
				ForProvider: smtpv1beta1.SMTPCredentialParameters{
					Domain: domainName,
					Login:  "test",
				},
			},
		}
		_, err = smtpExternal.Create(ctx, smtpMg)
		require.NoError(t, err)

		_, err = mockClient.CreateWebhook(ctx, domainName, &clients.WebhookSpec{
			EventType: "delivered",
			URL:       "https://example.com/webhook",
		})
		require.NoError(t, err)

		templateExternal := &template.ExternalForTesting{Client: mockClient}
		templateMg := &templatev1beta1.Template{
			Spec: templatev1beta1.TemplateSpec{
				ForProvider: templatev1beta1.TemplateParameters{
					Domain: domainName,
					Name:   "test-template",
				},
			},
		}
		_, err = templateExternal.Create(ctx, templateMg)
		require.NoError(t, err)
	})

	// Delete dependent resources first
	t.Run("DeleteDependentResources", func(t *testing.T) {
		// Delete SMTP credential
		smtpExternal := &smtpcredential.ExternalForTesting{Client: mockClient}
		smtpMg := &smtpv1beta1.SMTPCredential{
			Spec: smtpv1beta1.SMTPCredentialSpec{
				ForProvider: smtpv1beta1.SMTPCredentialParameters{
					Domain: domainName,
					Login:  "test",
				},
			},
		}
		_, err := smtpExternal.Delete(ctx, smtpMg)
		require.NoError(t, err)

		// Delete webhook
		err = mockClient.DeleteWebhook(ctx, domainName, "delivered")
		require.NoError(t, err)

		// Delete template
		templateExternal := &template.ExternalForTesting{Client: mockClient}
		templateMg := &templatev1beta1.Template{
			Spec: templatev1beta1.TemplateSpec{
				ForProvider: templatev1beta1.TemplateParameters{
					Domain: domainName,
					Name:   "test-template",
				},
			},
		}
		_, err = templateExternal.Delete(ctx, templateMg)
		require.NoError(t, err)
	})

	// Now delete the domain
	t.Run("DeleteDomain", func(t *testing.T) {
		domainExternal := &domain.ExternalForTesting{Client: mockClient}
		domainMg := &v1beta1.Domain{
			Spec: v1beta1.DomainSpec{
				ForProvider: v1beta1.DomainParameters{
					Name: domainName,
				},
			},
		}

		_, err := domainExternal.Delete(ctx, domainMg)
		require.NoError(t, err)

		// Verify domain is deleted
		_, err = mockClient.GetDomain(ctx, domainName)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "domain not found")
	})
}

// TestMultipleDomainScenario tests managing multiple domains with their resources
func TestMultipleDomainScenario(t *testing.T) {
	mockClient := NewIntegrationMockClient()
	ctx := context.Background()

	domains := []string{"example1.com", "example2.com", "example3.com"}

	// Create multiple domains with their own resources
	t.Run("CreateMultipleDomains", func(t *testing.T) {
		for i, domainName := range domains {
			// Create domain
			domainExternal := &domain.ExternalForTesting{Client: mockClient}
			domainMg := &v1beta1.Domain{
				Spec: v1beta1.DomainSpec{
					ForProvider: v1beta1.DomainParameters{
						Name: domainName,
					},
				},
			}
			_, err := domainExternal.Create(ctx, domainMg)
			require.NoError(t, err)

			// Create domain-specific mailing list
			mailingListExternal := &mailinglist.ExternalForTesting{Client: mockClient}
			mailingListMg := &mailinglistv1beta1.MailingList{
				Spec: mailinglistv1beta1.MailingListSpec{
					ForProvider: mailinglistv1beta1.MailingListParameters{
						Address: "newsletter@" + domainName,
						Name:    stringPtr("Newsletter " + domainName),
					},
				},
			}
			_, err = mailingListExternal.Create(ctx, mailingListMg)
			require.NoError(t, err)

			// Create domain-specific SMTP credential
			smtpExternal := &smtpcredential.ExternalForTesting{Client: mockClient}
			smtpMg := &smtpv1beta1.SMTPCredential{
				Spec: smtpv1beta1.SMTPCredentialSpec{
					ForProvider: smtpv1beta1.SMTPCredentialParameters{
						Domain: domainName,
						Login:  "admin" + string(rune('0'+i)),
					},
				},
			}
			_, err = smtpExternal.Create(ctx, smtpMg)
			require.NoError(t, err)
		}
	})

	// Verify all resources were created correctly
	t.Run("VerifyMultipleDomainSetup", func(t *testing.T) {
		for i, domainName := range domains {
			// Verify domain exists
			domain, err := mockClient.GetDomain(ctx, domainName)
			require.NoError(t, err)
			assert.Equal(t, domainName, domain.Name)

			// Verify mailing list exists
			mailingList, err := mockClient.GetMailingList(ctx, "newsletter@"+domainName)
			require.NoError(t, err)
			assert.Equal(t, "newsletter@"+domainName, mailingList.Address)

			// Verify SMTP credential exists
			smtp, err := mockClient.GetSMTPCredential(ctx, domainName, "admin"+string(rune('0'+i)))
			require.NoError(t, err)
			assert.Equal(t, "admin"+string(rune('0'+i)), smtp.Login)
		}

		// Verify total counts
		assert.Len(t, mockClient.domains, 3)
		assert.Len(t, mockClient.mailingLists, 3)
		assert.Len(t, mockClient.smtpCredentials, 3)
	})
}
