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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	webhooktypes "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
)

// Route Client Tests
func TestRouteOperations(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		operation func(client Client) error
		path      string
		status    int
		hasError  bool
	}{
		{
			name:   "CreateRoute success",
			method: "POST",
			operation: func(client Client) error {
				spec := &routetypes.RouteParameters{
					Priority:    intPtr(10),
					Description: stringPtr("Test route"),
					Expression:  "match_recipient(\"test@example.com\")",
					Actions:     []routetypes.RouteAction{{Type: "forward", Destination: stringPtr("admin@example.com")}},
				}
				_, err := client.CreateRoute(context.Background(), spec)
				return err
			},
			path:     "/v3/routes",
			status:   200,
			hasError: false,
		},
		{
			name:   "GetRoute success",
			method: "GET",
			operation: func(client Client) error {
				_, err := client.GetRoute(context.Background(), "route_123")
				return err
			},
			path:     "/v3/routes/route_123",
			status:   200,
			hasError: false,
		},
		{
			name:   "DeleteRoute success",
			method: "DELETE",
			operation: func(client Client) error {
				return client.DeleteRoute(context.Background(), "route_123")
			},
			path:     "/v3/routes/route_123",
			status:   200,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)

				w.WriteHeader(tt.status)
				if tt.method == "GET" || tt.method == "POST" {
					response := map[string]interface{}{
						"route": map[string]interface{}{
							"id":          "route_123",
							"priority":    10,
							"description": "Test route",
							"expression":  "match_recipient(\"test@example.com\")",
							"actions":     []map[string]interface{}{{"type": "forward", "destination": "admin@example.com"}},
							"created_at":  "2025-01-01T00:00:00Z",
						},
					}
					_ = json.NewEncoder(w).Encode(response)
				} else {
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "success"})
				}
			}))
			defer server.Close()

			config := &Config{APIKey: "test-key", BaseURL: server.URL + "/v3", HTTPClient: &http.Client{}}
			client := NewClient(config)

			err := tt.operation(client)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Webhook Client Tests
func TestWebhookOperations(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		operation func(client Client) error
		path      string
		status    int
		hasError  bool
	}{
		{
			name:   "CreateWebhook success",
			method: "POST",
			operation: func(client Client) error {
				spec := &webhooktypes.WebhookParameters{
					EventType: "delivered",
					URL:       "https://example.com/webhook",
					Username:  stringPtr("user"),
				}
				_, err := client.CreateWebhook(context.Background(), "example.com", spec)
				return err
			},
			path:     "/v3/domains/example.com/webhooks/delivered",
			status:   200,
			hasError: false,
		},
		{
			name:   "GetWebhook success",
			method: "GET",
			operation: func(client Client) error {
				_, err := client.GetWebhook(context.Background(), "example.com", "delivered")
				return err
			},
			path:     "/v3/domains/example.com/webhooks/delivered",
			status:   200,
			hasError: false,
		},
		{
			name:   "DeleteWebhook success",
			method: "DELETE",
			operation: func(client Client) error {
				return client.DeleteWebhook(context.Background(), "example.com", "delivered")
			},
			path:     "/v3/domains/example.com/webhooks/delivered",
			status:   200,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)

				w.WriteHeader(tt.status)
				if tt.method == "GET" || tt.method == "POST" {
					response := map[string]interface{}{
						"webhook": map[string]interface{}{
							"id":         "webhook_123",
							"event_type": "delivered",
							"url":        "https://example.com/webhook",
							"username":   "user",
							"created_at": "2025-01-01T00:00:00Z",
						},
					}
					_ = json.NewEncoder(w).Encode(response)
				} else {
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "success"})
				}
			}))
			defer server.Close()

			config := &Config{APIKey: "test-key", BaseURL: server.URL + "/v3", HTTPClient: &http.Client{}}
			client := NewClient(config)

			err := tt.operation(client)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// SMTP Credential Client Tests
func TestSMTPCredentialOperations(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		operation func(client Client) error
		path      string
		status    int
		hasError  bool
	}{
		{
			name:   "CreateSMTPCredential success",
			method: "POST",
			operation: func(client Client) error {
				spec := &smtpcredentialtypes.SMTPCredentialParameters{
					Login:    "test@example.com",
					Password: stringPtr("password123"),
				}
				_, err := client.CreateSMTPCredential(context.Background(), "example.com", spec)
				return err
			},
			path:     "/v3/domains/example.com/credentials",
			status:   200,
			hasError: false,
		},
		{
			name:   "GetSMTPCredential success",
			method: "GET",
			operation: func(client Client) error {
				_, err := client.GetSMTPCredential(context.Background(), "example.com", "test@example.com")
				return err
			},
			path:     "/v3/domains/example.com/credentials",
			status:   200,
			hasError: false,
		},
		{
			name:   "DeleteSMTPCredential success",
			method: "DELETE",
			operation: func(client Client) error {
				return client.DeleteSMTPCredential(context.Background(), "example.com", "test@example.com")
			},
			path:     "/v3/domains/example.com/credentials/test@example.com",
			status:   200,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle the specific CreateSMTPCredential flow (POST -> GET)
				if tt.name == "CreateSMTPCredential success" {
					if r.Method == "POST" && r.URL.Path == tt.path {
						w.WriteHeader(tt.status)
						// Return success message (current Mailgun API behavior)
						response := map[string]interface{}{
							"message": "Created 1 credentials pair(s)",
						}
						_ = json.NewEncoder(w).Encode(response)
						return
					} else if r.Method == "GET" && r.URL.Path == "/v3/domains/example.com/credentials" {
						w.WriteHeader(200)
						// Return credentials list with the newly created credential
						response := map[string]interface{}{
							"items": []map[string]interface{}{
								{
									"login":      "test@example.com",
									"password":   "password123", // Mailgun-generated password
									"created_at": "2025-01-01T00:00:00Z",
									"state":      "active",
								},
							},
						}
						_ = json.NewEncoder(w).Encode(response)
						return
					}
				}

				// For other operations, use the original logic
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)

				w.WriteHeader(tt.status)
				switch tt.method {
				case "GET":
					// GetSMTPCredential returns a list of credentials
					response := map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"login":      "test@example.com",
								"password":   "password123",
								"created_at": "2025-01-01T00:00:00Z",
								"state":      "active",
							},
						},
					}
					_ = json.NewEncoder(w).Encode(response)
				case "POST":
					// CreateSMTPCredential returns a success message
					response := map[string]interface{}{
						"message": "Created 1 credentials pair(s)",
					}
					_ = json.NewEncoder(w).Encode(response)
				default:
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "success"})
				}
			}))
			defer server.Close()

			config := &Config{APIKey: "test-key", BaseURL: server.URL + "/v3", HTTPClient: &http.Client{}}
			client := NewClient(config)

			err := tt.operation(client)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Template Client Tests
func TestTemplateOperations(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		operation func(client Client) error
		path      string
		status    int
		hasError  bool
	}{
		{
			name:   "CreateTemplate success",
			method: "POST",
			operation: func(client Client) error {
				spec := &templatetypes.TemplateParameters{
					Name:        "test-template",
					Description: stringPtr("Test template"),
					Template:    stringPtr("Hello {{name}}!"),
					Engine:      stringPtr("mustache"),
				}
				_, err := client.CreateTemplate(context.Background(), "example.com", spec)
				return err
			},
			path:     "/v3/domains/example.com/templates",
			status:   200,
			hasError: false,
		},
		{
			name:   "GetTemplate success",
			method: "GET",
			operation: func(client Client) error {
				_, err := client.GetTemplate(context.Background(), "example.com", "test-template")
				return err
			},
			path:     "/v3/domains/example.com/templates/test-template",
			status:   200,
			hasError: false,
		},
		{
			name:   "DeleteTemplate success",
			method: "DELETE",
			operation: func(client Client) error {
				return client.DeleteTemplate(context.Background(), "example.com", "test-template")
			},
			path:     "/v3/domains/example.com/templates/test-template",
			status:   200,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)

				w.WriteHeader(tt.status)
				if tt.method == "GET" || tt.method == "POST" {
					response := map[string]interface{}{
						"template": map[string]interface{}{
							"id":          "template_123",
							"name":        "test-template",
							"description": "Test template",
							"template":    "Hello {{name}}!",
							"engine":      "mustache",
							"created_at":  "2025-01-01T00:00:00Z",
						},
					}
					_ = json.NewEncoder(w).Encode(response)
				} else {
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "success"})
				}
			}))
			defer server.Close()

			config := &Config{APIKey: "test-key", BaseURL: server.URL + "/v3", HTTPClient: &http.Client{}}
			client := NewClient(config)

			err := tt.operation(client)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Error handling tests
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		operation  func(client Client) error
	}{
		{
			name:       "400 Bad Request",
			statusCode: 400,
			response:   "Bad Request: Invalid parameters",
			operation: func(client Client) error {
				_, err := client.CreateDomain(context.Background(), &domaintypes.DomainParameters{Name: ""})
				return err
			},
		},
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			response:   "Unauthorized: Invalid API key",
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "test.com")
				return err
			},
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			response:   "Not Found: Domain does not exist",
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "nonexistent.com")
				return err
			},
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			response:   "Internal Server Error",
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "error.com")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			config := &Config{APIKey: "test-key", BaseURL: server.URL + "/v3", HTTPClient: &http.Client{}}
			client := NewClient(config)

			err := tt.operation(client)
			require.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("%d", tt.statusCode)) // Check status code is in error
		})
	}
}

// Network and Connection Failure Tests
func TestNetworkFailures(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*Config, func())
		operation func(client Client) error
		wantErr   string
	}{
		{
			name: "Connection Refused",
			setupFunc: func() (*Config, func()) {
				// Use a port that's guaranteed to be closed
				config := &Config{
					APIKey:     "test-key",
					BaseURL:    "http://localhost:1", // Port 1 should be closed
					HTTPClient: &http.Client{},
				}
				return config, func() {}
			},
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "test.com")
				return err
			},
			wantErr: "connection refused",
		},
		{
			name: "DNS Resolution Failure",
			setupFunc: func() (*Config, func()) {
				config := &Config{
					APIKey:     "test-key",
					BaseURL:    "http://nonexistent-domain-12345.invalid/v3",
					HTTPClient: &http.Client{},
				}
				return config, func() {}
			},
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "test.com")
				return err
			},
			wantErr: "no such host",
		},
		{
			name: "Request Timeout",
			setupFunc: func() (*Config, func()) {
				// Create a server that never responds
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Sleep longer than our timeout
					time.Sleep(2 * time.Second)
					w.WriteHeader(200)
				}))

				// Use a very short timeout
				config := &Config{
					APIKey:  "test-key",
					BaseURL: server.URL + "/v3",
					HTTPClient: &http.Client{
						Timeout: 100 * time.Millisecond, // Very short timeout
					},
				}
				return config, server.Close
			},
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "test.com")
				return err
			},
			wantErr: "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, cleanup := tt.setupFunc()
			defer cleanup()

			client := NewClient(config)
			err := tt.operation(client)

			require.Error(t, err, "Expected network failure to return error")
			assert.Contains(t, strings.ToLower(err.Error()), tt.wantErr,
				"Error should contain expected failure type. Got: %v", err)
		})
	}
}

// Malformed Response Tests
func TestMalformedResponses(t *testing.T) {
	tests := []struct {
		name      string
		response  string
		operation func(client Client) error
		wantErr   string
	}{
		{
			name:     "Invalid JSON Response",
			response: "{invalid json content",
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "test.com")
				return err
			},
			wantErr: "invalid character",
		},
		{
			name:     "Empty Response Body",
			response: "",
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "test.com")
				return err
			},
			wantErr: "EOF", // JSON parsing error for empty body
		},
		{
			name:     "Partial JSON Response",
			response: `{"domain":{"name":"test.com","state":"partial"`,
			operation: func(client Client) error {
				_, err := client.GetDomain(context.Background(), "test.com")
				return err
			},
			wantErr: "unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			config := &Config{APIKey: "test-key", BaseURL: server.URL + "/v3", HTTPClient: &http.Client{}}
			client := NewClient(config)

			err := tt.operation(client)

			if tt.wantErr != "" {
				require.Error(t, err, "Expected malformed response to return error")
				assert.Contains(t, err.Error(), tt.wantErr,
					"Error should contain expected parsing failure. Got: %v", err)
			}
		})
	}
}

// Context Cancellation Tests
func TestContextCancellation(t *testing.T) {
	tests := []struct {
		name      string
		operation func(ctx context.Context, client Client) error
	}{
		{
			name: "GetDomain with Cancelled Context",
			operation: func(ctx context.Context, client Client) error {
				_, err := client.GetDomain(ctx, "test.com")
				return err
			},
		},
		{
			name: "CreateDomain with Cancelled Context",
			operation: func(ctx context.Context, client Client) error {
				_, err := client.CreateDomain(ctx, &domaintypes.DomainParameters{Name: "test.com"})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate a slow server
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"domain":{"name":"test.com"}}`))
			}))
			defer server.Close()

			config := &Config{APIKey: "test-key", BaseURL: server.URL + "/v3", HTTPClient: &http.Client{}}
			client := NewClient(config)

			// Create a context that we'll cancel immediately
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			err := tt.operation(ctx, client)
			require.Error(t, err, "Expected cancelled context to return error")
			assert.Contains(t, err.Error(), "context canceled",
				"Error should indicate context cancellation. Got: %v", err)
		})
	}
}
