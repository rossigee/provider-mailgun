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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				spec := &RouteSpec{
					Priority:    intPtr(10),
					Description: stringPtr("Test route"),
					Expression:  "match_recipient(\"test@example.com\")",
					Actions:     []RouteAction{{Type: "forward", Destination: stringPtr("admin@example.com")}},
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
				spec := &WebhookSpec{
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
				spec := &SMTPCredentialSpec{
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
					// CreateSMTPCredential returns single credential
					response := map[string]interface{}{
						"login":      "test@example.com",
						"password":   "password123",
						"created_at": "2025-01-01T00:00:00Z",
						"state":      "active",
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
				spec := &TemplateSpec{
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
				_, err := client.CreateDomain(context.Background(), &DomainSpec{Name: ""})
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
