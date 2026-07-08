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

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	bouncetypes "github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglisttypes "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	v1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	webhooktypes "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error)
	GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error)
	UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error)
	DeleteDomain(ctx context.Context, name string) error

	// MailingList operations
	CreateMailingList(ctx context.Context, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error)
	GetMailingList(ctx context.Context, address string) (*mailinglisttypes.MailingListObservation, error)
	UpdateMailingList(ctx context.Context, address string, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error)
	DeleteMailingList(ctx context.Context, address string) error

	// Route operations
	CreateRoute(ctx context.Context, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error)
	GetRoute(ctx context.Context, id string) (*routetypes.RouteObservation, error)
	UpdateRoute(ctx context.Context, id string, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error)
	DeleteRoute(ctx context.Context, id string) error

	// Webhook operations
	CreateWebhook(ctx context.Context, domain string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error)
	GetWebhook(ctx context.Context, domain, eventType string) (*webhooktypes.WebhookObservation, error)
	UpdateWebhook(ctx context.Context, domain, eventType string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error)
	DeleteWebhook(ctx context.Context, domain, eventType string) error

	// SMTPCredential operations
	CreateSMTPCredential(ctx context.Context, domain string, credential *smtpcredentialtypes.SMTPCredentialParameters) (*smtpcredentialtypes.SMTPCredentialObservation, error)
	GetSMTPCredential(ctx context.Context, domain, login string) (*smtpcredentialtypes.SMTPCredentialObservation, error)
	UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*smtpcredentialtypes.SMTPCredentialObservation, error)
	DeleteSMTPCredential(ctx context.Context, domain, login string) error

	// Template operations
	CreateTemplate(ctx context.Context, domain string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error)
	GetTemplate(ctx context.Context, domain, name string) (*templatetypes.TemplateObservation, error)
	UpdateTemplate(ctx context.Context, domain, name string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error)
	DeleteTemplate(ctx context.Context, domain, name string) error

	// Bounce suppression operations
	CreateBounce(ctx context.Context, domain string, bounce *bouncetypes.BounceParameters) (*bouncetypes.BounceObservation, error)
	GetBounce(ctx context.Context, domain, address string) (*bouncetypes.BounceObservation, error)
	DeleteBounce(ctx context.Context, domain, address string) error

	// Complaint suppression operations (temporarily using interface until types exist)
	CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error)
	GetComplaint(ctx context.Context, domain, address string) (interface{}, error)
	DeleteComplaint(ctx context.Context, domain, address string) error

	// Unsubscribe suppression operations (temporarily using interface until types exist)
	CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error)
	GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error)
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
		transport := &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false, // Enable keep-alives with proper connection management
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  false,
			MaxIdleConnsPerHost: 2,    // Limit concurrent connections per host
			ForceAttemptHTTP2:   true, // Enable HTTP/2 which works better with Mailgun
		}
		config.HTTPClient = &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		}
	}
	return &mailgunClient{config: config}
}

// getProviderConfigReference extracts the provider config reference from a managed resource
func getProviderConfigReference(mg resource.Managed) *xpv1.ProviderConfigReference {
	// Type switch to handle different resource types and access their ProviderConfigReference
	switch v := mg.(type) {
	case interface {
		GetProviderConfigReference() *xpv1.ProviderConfigReference
	}:
		return v.GetProviderConfigReference()
	default:
		// If we can't determine the provider config reference, return nil
		return nil
	}
}

// GetConfig extracts the configuration from the provider config
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	// Get the provider config reference from the managed resource's spec
	if pcRef := getProviderConfigReference(mg); pcRef != nil {
		return UseProviderConfig(ctx, c, mg, pcRef)
	}

	return nil, errors.New("no credentials specified")
}

// UseProviderConfig extracts configuration from a ProviderConfig
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed, pcRef *xpv1.ProviderConfigReference) (*Config, error) {
	pc := &v1beta1.ProviderConfig{}

	// For cluster-scoped resources, we need to look in a default namespace
	// For namespaced resources, look in the same namespace as the resource
	namespace := mg.GetNamespace()
	if namespace == "" {
		// Cluster-scoped resource - look for ProviderConfig in crossplane-system by default
		namespace = "crossplane-system"
	}

	if err := c.Get(ctx, types.NamespacedName{
		Name:      pcRef.Name,
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

	// Store the original body data for retries
	var originalBodyData []byte
	if body != nil {
		if data, err := io.ReadAll(body); err == nil {
			originalBodyData = data
		}
	}

	// Create initial request body reader from stored data
	var requestBody io.Reader
	if originalBodyData != nil {
		requestBody = strings.NewReader(string(originalBodyData))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.SetBasicAuth("api", c.config.APIKey)
	req.Header.Set("User-Agent", "crossplane-provider-mailgun")

	if originalBodyData != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Retry logic for 502 Bad Gateway errors
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			waitTime := time.Duration(attempt) * 2 * time.Second
			time.Sleep(waitTime)

			// Recreate the request for retry using stored body data
			var retryBody io.Reader
			if originalBodyData != nil {
				retryBody = strings.NewReader(string(originalBodyData))
			}
			req, err = http.NewRequestWithContext(ctx, method, url, retryBody)
			if err != nil {
				return nil, errors.Wrap(err, "failed to recreate request for retry")
			}
			req.SetBasicAuth("api", c.config.APIKey)
			req.Header.Set("User-Agent", "crossplane-provider-mailgun")
			if originalBodyData != nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}

		resp, err = c.config.HTTPClient.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return nil, errors.Wrap(err, "failed to execute request after retries")
			}
			continue
		}

		// If it's not a 502, return the response (success or other error)
		if resp.StatusCode != 502 {
			return resp, nil
		}

		// If it's a 502 and we have retries left, close this response and try again
		if attempt < maxRetries {
			_ = resp.Body.Close()
			continue
		}

		// Max retries reached with 502, return the last response
		return resp, nil
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
