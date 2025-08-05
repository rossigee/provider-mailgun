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

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rossigee/provider-mailgun/apis/v1beta1"
)

const (
	// DefaultBaseURL is the default Mailgun API base URL for US region
	DefaultBaseURL = "https://api.mailgun.net/v3"
	// EUBaseURL is the Mailgun API base URL for EU region
	EUBaseURL = "https://api.eu.mailgun.net/v3"

	// HTTP timeout for API requests
	defaultTimeout = 30 * time.Second
)

// Client interface for Mailgun API operations
type Client interface {
	// Domain operations
	CreateDomain(ctx context.Context, domain *DomainSpec) (*Domain, error)
	GetDomain(ctx context.Context, name string) (*Domain, error)
	UpdateDomain(ctx context.Context, name string, domain *DomainSpec) (*Domain, error)
	DeleteDomain(ctx context.Context, name string) error

	// MailingList operations
	CreateMailingList(ctx context.Context, list *MailingListSpec) (*MailingList, error)
	GetMailingList(ctx context.Context, address string) (*MailingList, error)
	UpdateMailingList(ctx context.Context, address string, list *MailingListSpec) (*MailingList, error)
	DeleteMailingList(ctx context.Context, address string) error

	// Route operations
	CreateRoute(ctx context.Context, route *RouteSpec) (*Route, error)
	GetRoute(ctx context.Context, id string) (*Route, error)
	UpdateRoute(ctx context.Context, id string, route *RouteSpec) (*Route, error)
	DeleteRoute(ctx context.Context, id string) error

	// Webhook operations
	CreateWebhook(ctx context.Context, domain string, webhook *WebhookSpec) (*Webhook, error)
	GetWebhook(ctx context.Context, domain, eventType string) (*Webhook, error)
	UpdateWebhook(ctx context.Context, domain, eventType string, webhook *WebhookSpec) (*Webhook, error)
	DeleteWebhook(ctx context.Context, domain, eventType string) error

	// SMTPCredential operations
	CreateSMTPCredential(ctx context.Context, domain string, credential *SMTPCredentialSpec) (*SMTPCredential, error)
	GetSMTPCredential(ctx context.Context, domain, login string) (*SMTPCredential, error)
	UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*SMTPCredential, error)
	DeleteSMTPCredential(ctx context.Context, domain, login string) error

	// Template operations
	CreateTemplate(ctx context.Context, domain string, template *TemplateSpec) (*Template, error)
	GetTemplate(ctx context.Context, domain, name string) (*Template, error)
	UpdateTemplate(ctx context.Context, domain, name string, template *TemplateSpec) (*Template, error)
	DeleteTemplate(ctx context.Context, domain, name string) error

	// Bounce suppression operations
	CreateBounce(ctx context.Context, domain string, bounce *BounceSpec) (*Bounce, error)
	GetBounce(ctx context.Context, domain, address string) (*Bounce, error)
	DeleteBounce(ctx context.Context, domain, address string) error

	// Complaint suppression operations
	CreateComplaint(ctx context.Context, domain string, complaint *ComplaintSpec) (*Complaint, error)
	GetComplaint(ctx context.Context, domain, address string) (*Complaint, error)
	DeleteComplaint(ctx context.Context, domain, address string) error

	// Unsubscribe suppression operations
	CreateUnsubscribe(ctx context.Context, domain string, unsubscribe *UnsubscribeSpec) (*Unsubscribe, error)
	GetUnsubscribe(ctx context.Context, domain, address string) (*Unsubscribe, error)
	DeleteUnsubscribe(ctx context.Context, domain, address string) error
}

// Config holds the configuration for the Mailgun client
type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// Credentials represents the structure of the credentials secret
type Credentials struct {
	APIKey string `json:"api_key"`
	Region string `json:"region,omitempty"`
}

// mailgunClient implements the Client interface
type mailgunClient struct {
	config *Config
}

// NewClient creates a new Mailgun client
func NewClient(config *Config) Client {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}
	return &mailgunClient{config: config}
}

// GetConfig extracts the configuration from the provider config
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	default:
		return nil, errors.New("no credentials specified")
	}
}

// UseProviderConfig extracts configuration from a ProviderConfig
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	pc := &v1beta1.ProviderConfig{}

	// For cluster-scoped resources, we need to look in a default namespace
	// For namespaced resources, look in the same namespace as the resource
	namespace := mg.GetNamespace()
	if namespace == "" {
		// Cluster-scoped resource - look for ProviderConfig in crossplane-system by default
		namespace = "crossplane-system"
	}

	if err := c.Get(ctx, types.NamespacedName{
		Name:      mg.GetProviderConfigReference().Name,
		Namespace: namespace,
	}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced ProviderConfig")
	}

	// Note: ProviderConfig usage tracking is optional

	data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, c, pc.Spec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials")
	}

	// Try to parse as JSON first (new format)
	var creds Credentials
	var apiKey string
	if err := json.Unmarshal(data, &creds); err == nil && creds.APIKey != "" {
		// JSON format with api_key field
		apiKey = creds.APIKey
	} else {
		// Fall back to treating the entire data as the API key (legacy format)
		apiKey = strings.TrimSpace(string(data))
		if apiKey == "" {
			return nil, errors.New("mailgun API key not found in credentials")
		}
	}

	if apiKey == "" {
		return nil, errors.New("mailgun API key not found in credentials")
	}

	baseURL := DefaultBaseURL
	if pc.Spec.APIBaseURL != nil {
		baseURL = *pc.Spec.APIBaseURL
	} else if pc.Spec.Region != nil && *pc.Spec.Region == "EU" {
		baseURL = EUBaseURL
	}

	return &Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}, nil
}

// Helper method to make HTTP requests
func (c *mailgunClient) makeRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.config.BaseURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.SetBasicAuth("api", c.config.APIKey)
	req.Header.Set("User-Agent", "crossplane-provider-mailgun")

	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}

	return resp, nil
}

// Helper method to handle API responses
func (c *mailgunClient) handleResponse(resp *http.Response, target interface{}) error {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return errors.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}

	return nil
}

// Helper method to create form data
func createFormData(params map[string]interface{}) string {
	values := url.Values{}
	for key, value := range params {
		if value != nil {
			values.Add(key, fmt.Sprintf("%v", value))
		}
	}
	return values.Encode()
}

// IsNotFound checks if an error represents a "not found" condition
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found")
}
