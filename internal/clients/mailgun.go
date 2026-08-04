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
	"strconv"
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

// Region describes a Mailgun regional deployment. Each entry maps a
// ProviderConfig.Spec.Region code ("US", "EU") to the corresponding API
// endpoints and SMTP relay hostname. Append a new entry here when Mailgun
// launches a new region; no other code change is required (the
// ProviderConfig CRD enum will also need to be regenerated).
type Region struct {
	// Code is the canonical short name used in ProviderConfig.Spec.Region.
	Code string
	// BaseURL is the Mailgun v3 API base URL for this region.
	BaseURL string
	// V4BaseURL is the Mailgun v4 Domains API base URL for this region.
	V4BaseURL string
	// SMTPHost is the SMTP relay hostname SMTP clients (Keycloak, Postfix,
	// …) connect to when authenticating with credentials issued from this
	// region.
	SMTPHost string
	// URLMarkers are substrings of an apiBaseURL that identify this region
	// when the user supplies an explicit apiBaseURL without setting
	// spec.region. The more specific marker should be listed first inside a
	// single region (e.g. "eu.mailgun.net" before "mailgun.net") so the
	// longest-prefix match wins during detection.
	URLMarkers []string
}

// regions is the registered list of Mailgun regions. Append here when a
// new region is added.
var regions = []Region{
	{
		Code:       "US",
		BaseURL:    "https://api.mailgun.net/v3",
		V4BaseURL:  "https://api.mailgun.net/v4",
		SMTPHost:   "smtp.mailgun.org",
		URLMarkers: []string{"mailgun.net"},
	},
	{
		Code:       "EU",
		BaseURL:    "https://api.eu.mailgun.net/v3",
		V4BaseURL:  "https://api.eu.mailgun.net/v4",
		SMTPHost:   "smtp.eu.mailgun.org",
		URLMarkers: []string{"eu.mailgun.net"},
	},
}

const (
	// DefaultSMTPHost is the SMTP relay hostname for the implicit default
	// region (the first entry of the regions registry, currently US).
	// Exposed as a constant for backwards-compat with callers that resolve
	// the default without a ProviderConfig in hand. The first region entry
	// in `regions` below must use this same value.
	DefaultSMTPHost = "smtp.mailgun.org"

	// HTTP timeout for API requests
	defaultTimeout = 30 * time.Second
)

// findRegionByCode looks up a region by its Code. Returns false when the
// code is not registered.
func findRegionByCode(code string) (Region, bool) {
	for _, r := range regions {
		if r.Code == code {
			return r, true
		}
	}
	return Region{}, false
}

// detectRegionByURL returns the region whose URLMarkers appear in baseURL.
// When more than one region's marker matches (e.g. "eu.mailgun.net" matches
// both EU and the more general US "mailgun.net" marker), the longest
// matching marker wins so the most specific region is selected.
// Returns false when no marker matches.
func detectRegionByURL(baseURL string) (Region, bool) {
	var best Region
	bestLen := -1
	matched := false
	for _, r := range regions {
		for _, marker := range r.URLMarkers {
			if !strings.Contains(baseURL, marker) {
				continue
			}
			if len(marker) > bestLen {
				best = r
				bestLen = len(marker)
				matched = true
			}
		}
	}
	return best, matched
}

// DeriveV4BaseURL returns the v4 Domains API base URL corresponding to a
// v3 (or unspecified-suffix) base URL. If baseURL ends with /v3 the suffix
// is swapped for /v4; an explicit /v4 is kept verbatim; otherwise /v4 is
// appended. Exported so callers that build Config manually (e.g. the
// health check) can populate V4BaseURL without duplicating the rule.
func DeriveV4BaseURL(baseURL string) string {
	switch {
	case strings.HasSuffix(baseURL, "/v3"):
		return strings.TrimSuffix(baseURL, "/v3") + "/v4"
	case strings.HasSuffix(baseURL, "/v4"):
		return baseURL
	default:
		return baseURL + "/v4"
	}
}

// Client interface for Mailgun API operations
type Client interface {
	// Domain operations
	CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error)
	GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error)
	UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error)
	DeleteDomain(ctx context.Context, name string) error
	VerifyDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error)

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
	APIKey   string
	BaseURL  string
	// V4BaseURL is the base URL for Mailgun v4 Domains API endpoints. When
	// the user supplies a BaseURL with a trailing /v3 (the historic default),
	// V4BaseURL is derived by swapping /v3 for /v4 so domain management calls
	// can hit /v4/domains, /v4/domains/{name} and /v4/domains/{name}/verify.
	V4BaseURL string
	// SMTPHost is the hostname SMTP clients should use to deliver mail
	// through this ProviderConfig. It is derived from the region
	// (US→smtp.mailgun.org, EU→smtp.eu.mailgun.org) and propagated into the
	// connection secret so downstream consumers (Keycloak, Odoo, …) get a
	// host that matches where the credential actually authenticates.
	SMTPHost  string
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

	// Select the region. Priority: explicit spec.region (if recognised) →
	// URL marker on apiBaseURL → first registered region (US).
	region := regions[0]
	if pc.Spec.Region != "" {
		if r, ok := findRegionByCode(pc.Spec.Region); ok {
			region = r
		}
	}
	if pc.Spec.APIBaseURL != nil {
		if r, ok := detectRegionByURL(*pc.Spec.APIBaseURL); ok {
			region = r
		}
	}

	// Resolve base URL. If the caller supplied an explicit apiBaseURL we
	// honour it verbatim; otherwise use the region default.
	baseURL := region.BaseURL
	if pc.Spec.APIBaseURL != nil {
		baseURL = *pc.Spec.APIBaseURL
	}

	// Derive v4 base URL. If baseURL ends with /v3 (historic default),
	// swap for /v4; otherwise honour an explicit /v4 suffix, or append /v4.
	v4BaseURL := region.V4BaseURL
	if pc.Spec.APIBaseURL != nil {
		v4BaseURL = DeriveV4BaseURL(baseURL)
	}

	// Derive SMTP host. The relay hostname is what SMTP clients (Keycloak,
	// Postfix, …) connect to; it must match the region of the credentials.
	// Region is authoritative when set; APIBaseURL-based detection is used
	// when Region is unset so an explicit apiBaseURL still produces the
	// matching SMTP relay.
	smtpHost := region.SMTPHost

	return &Config{
		APIKey:    apiKey,
		BaseURL:   baseURL,
		V4BaseURL: v4BaseURL,
		SMTPHost:  smtpHost,
	}, nil
}

// Helper method to make HTTP requests
func (c *mailgunClient) makeRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.makeRequestAt(ctx, c.config.BaseURL, method, path, body)
}

// RateLimitError is returned when the Mailgun API responds with HTTP 429
// Too Many Requests. It carries the retry duration Mailgun (or the
// upstream proxy) asked for, either parsed from the Retry-After header or
// derived from the X-RateLimit-Reset epoch.
type RateLimitError struct {
	// RetryAfter is the duration to wait before issuing the next request.
	RetryAfter time.Duration
	// Reset is the time at which the rate-limit window resets, when the
	// X-RateLimit-Reset header was present.
	Reset time.Time
	// Limit and Remaining mirror the X-RateLimit-* informational headers
	// when present.
	Limit     string
	Remaining string
}

// Error renders a human-readable description of the rate-limit event.
func (e *RateLimitError) Error() string {
	parts := []string{"mailgun API rate limit exceeded (HTTP 429)"}
	if e.RetryAfter > 0 {
		parts = append(parts, fmt.Sprintf("retry after %s", e.RetryAfter))
	}
	if !e.Reset.IsZero() {
		parts = append(parts, fmt.Sprintf("reset at %s", e.Reset.Format(time.RFC3339)))
	}
	if e.Limit != "" {
		parts = append(parts, fmt.Sprintf("limit=%s remaining=%s", e.Limit, e.Remaining))
	}
	return strings.Join(parts, "; ")
}

// parseRateLimit extracts Retry-After / X-RateLimit-* hints from a 429
// response. Returns nil for any other status code.
func parseRateLimit(resp *http.Response) *RateLimitError {
	if resp.StatusCode != http.StatusTooManyRequests {
		return nil
	}
	e := &RateLimitError{
		Limit:     resp.Header.Get("X-RateLimit-Limit"),
		Remaining: resp.Header.Get("X-RateLimit-Remaining"),
	}

	// Retry-After may be a number of seconds or an HTTP-date.
	if h := resp.Header.Get("Retry-After"); h != "" {
		if secs, err := strconv.Atoi(h); err == nil {
			e.RetryAfter = time.Duration(secs) * time.Second
		} else if t, err := http.ParseTime(h); err == nil {
			d := time.Until(t)
			if d > 0 {
				e.RetryAfter = d
			}
		}
	}

	// X-RateLimit-Reset is a unix epoch (seconds) in Mailgun's response.
	if h := resp.Header.Get("X-RateLimit-Reset"); h != "" {
		if epoch, err := strconv.ParseInt(h, 10, 64); err == nil {
			e.Reset = time.Unix(epoch, 0)
		}
	}

	if e.RetryAfter == 0 && !e.Reset.IsZero() {
		if d := time.Until(e.Reset); d > 0 {
			e.RetryAfter = d
		}
	}

	return e
}

// makeRequestAt issues a request against an explicit base URL, used by v4
// Domains API calls.
func (c *mailgunClient) makeRequestAt(ctx context.Context, baseURL, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", baseURL, path)

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

	// Retry logic handles transient 502 Bad Gateway and 429 Too Many Requests
	// responses. 429 honours Mailgun's Retry-After / X-RateLimit-Reset hints
	// so we back off exactly as long as Mailgun asks.
	var resp *http.Response
	var lastRateLimit *RateLimitError
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			var waitTime time.Duration
			if lastRateLimit != nil && lastRateLimit.RetryAfter > 0 {
				// Respect server-supplied back-off. Cap to 5 minutes so a
				// misbehaving upstream cannot wedge a reconcile forever.
				waitTime = lastRateLimit.RetryAfter
				if waitTime > 5*time.Minute {
					waitTime = 5 * time.Minute
				}
			} else {
				waitTime = time.Duration(attempt) * 2 * time.Second
			}

			// Honour the request context while sleeping.
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

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

		// Handle 429 Too Many Requests before any other status code. We
		// honour the server-supplied Retry-After / X-RateLimit-Reset.
		if rlErr := parseRateLimit(resp); rlErr != nil {
			_ = resp.Body.Close()
			lastRateLimit = rlErr
			if attempt < maxRetries {
				continue
			}
			return nil, rlErr
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
