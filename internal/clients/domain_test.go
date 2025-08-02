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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainSpec     *DomainSpec
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedDomain *Domain
		expectedError  bool
	}{
		{
			name: "successful creation with minimal params",
			domainSpec: &DomainSpec{
				Name: "test.com",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v3/domains", r.URL.Path)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

				// Verify request body
				_ = r.ParseForm()
				assert.Equal(t, "test.com", r.FormValue("name"))

				w.WriteHeader(http.StatusOK)
				response := map[string]interface{}{
					"domain": map[string]interface{}{
						"name":         "test.com",
						"type":         "sending",
						"state":        "unverified",
						"created_at":   "2025-01-01T00:00:00Z",
						"smtp_login":   "postmaster@test.com",
						"smtp_password": "generated-password",
						"required_dns_records": []map[string]interface{}{
							{
								"name":     "test.com",
								"record_type": "TXT",
								"value":    "v=spf1 include:mailgun.org ~all",
								"priority": nil,
								"valid":    false,
							},
						},
					},
				}
				_ = json.NewEncoder(w).Encode(response)
			},
			expectedDomain: &Domain{
				Name:         "test.com",
				Type:         "sending",
				State:        "unverified",
				CreatedAt:    "2025-01-01T00:00:00Z",
				SMTPLogin:    "postmaster@test.com",
				SMTPPassword: "generated-password",
				RequiredDNSRecords: []DNSRecord{
					{
						Name:     "test.com",
						Type:     "TXT",
						Value:    "v=spf1 include:mailgun.org ~all",
						Priority: nil,
						Valid:    boolPtr(false), // "unknown" -> false
					},
				},
			},
			expectedError: false,
		},
		{
			name: "successful creation with all params",
			domainSpec: &DomainSpec{
				Name:               "full.com",
				Type:               stringPtr("receiving"),
				ForceDKIMAuthority: boolPtr(true),
				DKIMKeySize:        intPtr(2048),
				SMTPPassword:       stringPtr("custom-password"),
				SpamAction:         stringPtr("block"),
				WebScheme:          stringPtr("https"),
				Wildcard:           boolPtr(true),
				IPs:                []string{"192.168.1.1", "192.168.1.2"},
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Verify all params are sent
				_ = r.ParseForm()
				assert.Equal(t, "full.com", r.FormValue("name"))
				assert.Equal(t, "receiving", r.FormValue("type"))
				assert.Equal(t, "true", r.FormValue("force_dkim_authority"))
				assert.Equal(t, "2048", r.FormValue("dkim_key_size"))
				assert.Equal(t, "custom-password", r.FormValue("smtp_password"))
				assert.Equal(t, "block", r.FormValue("spam_action"))
				assert.Equal(t, "https", r.FormValue("web_scheme"))
				assert.Equal(t, "true", r.FormValue("wildcard"))
				assert.Equal(t, "192.168.1.1,192.168.1.2", r.FormValue("ips"))

				w.WriteHeader(http.StatusOK)
				response := map[string]interface{}{
					"domain": map[string]interface{}{
						"name":  "full.com",
						"type":  "receiving",
						"state": "active",
					},
				}
				_ = json.NewEncoder(w).Encode(response)
			},
			expectedDomain: &Domain{
				Name:  "full.com",
				Type:  "receiving",
				State: "active",
			},
			expectedError: false,
		},
		{
			name: "server error",
			domainSpec: &DomainSpec{
				Name: "error.com",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Domain already exists"))
			},
			expectedDomain: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create client
			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			// Test CreateDomain
			result, err := client.CreateDomain(context.Background(), tt.domainSpec)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDomain, result)
			}
		})
	}
}

func TestGetDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainName     string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedDomain *Domain
		expectedError  bool
	}{
		{
			name:       "successful get",
			domainName: "example.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v3/domains/example.com", r.URL.Path)

				w.WriteHeader(http.StatusOK)
				response := map[string]interface{}{
					"domain": map[string]interface{}{
						"name":         "example.com",
						"type":         "sending",
						"state":        "active",
						"created_at":   "2025-01-01T00:00:00Z",
						"smtp_login":   "postmaster@example.com",
						"smtp_password": "password123",
					},
				}
				_ = json.NewEncoder(w).Encode(response)
			},
			expectedDomain: &Domain{
				Name:         "example.com",
				Type:         "sending",
				State:        "active",
				CreatedAt:    "2025-01-01T00:00:00Z",
				SMTPLogin:    "postmaster@example.com",
				SMTPPassword: "password123",
			},
			expectedError: false,
		},
		{
			name:       "domain not found",
			domainName: "notfound.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Domain not found"))
			},
			expectedDomain: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create client
			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			// Test GetDomain
			result, err := client.GetDomain(context.Background(), tt.domainName)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDomain, result)
			}
		})
	}
}

func TestUpdateDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainName     string
		domainSpec     *DomainSpec
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedDomain *Domain
		expectedError  bool
	}{
		{
			name:       "successful update",
			domainName: "update.com",
			domainSpec: &DomainSpec{
				SpamAction: stringPtr("tag"),
				WebScheme:  stringPtr("https"),
				Wildcard:   boolPtr(false),
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, "/v3/domains/update.com", r.URL.Path)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

				// Verify request body
				_ = r.ParseForm()
				assert.Equal(t, "tag", r.FormValue("spam_action"))
				assert.Equal(t, "https", r.FormValue("web_scheme"))
				assert.Equal(t, "false", r.FormValue("wildcard"))

				w.WriteHeader(http.StatusOK)
				response := map[string]interface{}{
					"domain": map[string]interface{}{
						"name":  "update.com",
						"type":  "sending",
						"state": "active",
					},
				}
				_ = json.NewEncoder(w).Encode(response)
			},
			expectedDomain: &Domain{
				Name:  "update.com",
				Type:  "sending",
				State: "active",
			},
			expectedError: false,
		},
		{
			name:       "domain not found",
			domainName: "notfound.com",
			domainSpec: &DomainSpec{
				SpamAction: stringPtr("disabled"),
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Domain not found"))
			},
			expectedDomain: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create client
			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			// Test UpdateDomain
			result, err := client.UpdateDomain(context.Background(), tt.domainName, tt.domainSpec)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDomain, result)
			}
		})
	}
}

func TestDeleteDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainName     string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
	}{
		{
			name:       "successful delete",
			domainName: "delete.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Equal(t, "/v3/domains/delete.com", r.URL.Path)

				w.WriteHeader(http.StatusOK)
				response := map[string]interface{}{
					"message": "Domain has been deleted",
				}
				_ = json.NewEncoder(w).Encode(response)
			},
			expectedError: false,
		},
		{
			name:       "domain not found",
			domainName: "notfound.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Domain not found"))
			},
			expectedError: true,
		},
		{
			name:       "server error",
			domainName: "error.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal server error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create client
			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			// Test DeleteDomain
			err := client.DeleteDomain(context.Background(), tt.domainName)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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

func intPtr(i int) *int {
	return &i
}
