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

	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
)

// stringPtr returns a pointer to the given string.
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to the given int.
func intPtr(i int) *int {
	return &i
}

// boolPtr returns a pointer to the given bool.
func boolPtr(b bool) *bool {
	return &b
}

func TestCreateDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainSpec     *domaintypes.DomainParameters
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedDomain *domaintypes.DomainObservation
		expectedError  bool
	}{
		{
			name: "successful creation with minimal params and DNS records at top level",
			domainSpec: &domaintypes.DomainParameters{
				Name: "test.com",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v4/domains", r.URL.Path)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

				_ = r.ParseForm()
				assert.Equal(t, "test.com", r.FormValue("name"))

				w.WriteHeader(http.StatusOK)
				// Mailgun v4 response: domain object plus DNS records at the top level
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"domain": map[string]interface{}{
						"name":          "test.com",
						"type":          "sending",
						"state":         "unverified",
						"created_at":    "2025-01-01T00:00:00Z",
						"smtp_login":    "postmaster@test.com",
						"smtp_password": "generated-password",
					},
					"receiving_dns_records": []map[string]interface{}{
						{
							"name":        "test.com",
							"record_type": "MX",
							"value":       "mxa.mailgun.org",
							"priority":    "10",
							"valid":       "unknown",
						},
					},
					"sending_dns_records": []map[string]interface{}{
						{
							"name":        "test.com",
							"record_type": "TXT",
							"value":       "v=spf1 include:mailgun.org ~all",
							"valid":       "unknown",
						},
					},
				})
			},
			expectedDomain: &domaintypes.DomainObservation{
				ID:           "test.com",
				State:        "unverified",
				CreatedAt:    "2025-01-01T00:00:00Z",
				SMTPLogin:    "postmaster@test.com",
				SMTPPassword: "generated-password",
				DNSVerified:  boolPtr(false), // at least one record is "unknown"
				ReceivingDNSRecords: []domaintypes.DNSRecord{
					{
						Name:     "test.com",
						Type:     "MX",
						Value:    "mxa.mailgun.org",
						Priority: stringPtr("10"),
						Valid:    stringPtr("unknown"),
					},
				},
				SendingDNSRecords: []domaintypes.DNSRecord{
					{
						Name:  "test.com",
						Type:  "TXT",
						Value: "v=spf1 include:mailgun.org ~all",
						Valid: stringPtr("unknown"),
					},
				},
			},
			expectedError: false,
		},
		{
			name: "successful creation with all params and all records valid",
			domainSpec: &domaintypes.DomainParameters{
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
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"domain": map[string]interface{}{
						"name":  "full.com",
						"type":  "receiving",
						"state": "active",
					},
					"sending_dns_records": []map[string]interface{}{
						{"name": "full.com", "record_type": "TXT", "value": "v=spf1", "valid": "valid"},
					},
				})
			},
			expectedDomain: &domaintypes.DomainObservation{
				ID:    "full.com",
				State: "active",
				DNSVerified: boolPtr(true),
				SendingDNSRecords: []domaintypes.DNSRecord{
					{Name: "full.com", Type: "TXT", Value: "v=spf1", Valid: stringPtr("valid")},
				},
			},
			expectedError: false,
		},
		{
			name: "creation with empty DNS records yields nil DNSVerified",
			domainSpec: &domaintypes.DomainParameters{
				Name: "empty.com",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"domain": map[string]interface{}{
						"name":  "empty.com",
						"state": "unverified",
					},
				})
			},
			expectedDomain: &domaintypes.DomainObservation{
				ID:          "empty.com",
				State:       "unverified",
				DNSVerified: nil,
			},
			expectedError: false,
		},
		{
			name: "server error",
			domainSpec: &domaintypes.DomainParameters{
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
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

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
		expectedDomain *domaintypes.DomainObservation
		expectedError  bool
	}{
		{
			name:       "successful get with DNS records",
			domainName: "example.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v4/domains/example.com", r.URL.Path)

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"domain": map[string]interface{}{
						"name":          "example.com",
						"type":          "sending",
						"state":         "active",
						"created_at":    "2025-01-01T00:00:00Z",
						"smtp_login":    "postmaster@example.com",
						"smtp_password": "password123",
					},
					"receiving_dns_records": []map[string]interface{}{
						{"name": "example.com", "record_type": "MX", "value": "mxa.mailgun.org", "priority": "10", "valid": "valid"},
					},
					"sending_dns_records": []map[string]interface{}{
						{"name": "example.com", "record_type": "TXT", "value": "v=spf1 include:mailgun.org ~all", "valid": "valid"},
						{"name": "k1._domainkey.example.com", "record_type": "TXT", "value": "k=rsa; p=...", "valid": "valid"},
					},
				})
			},
			expectedDomain: &domaintypes.DomainObservation{
				ID:           "example.com",
				State:        "active",
				CreatedAt:    "2025-01-01T00:00:00Z",
				SMTPLogin:    "postmaster@example.com",
				SMTPPassword: "password123",
				DNSVerified:  boolPtr(true),
				ReceivingDNSRecords: []domaintypes.DNSRecord{
					{Name: "example.com", Type: "MX", Value: "mxa.mailgun.org", Priority: stringPtr("10"), Valid: stringPtr("valid")},
				},
				SendingDNSRecords: []domaintypes.DNSRecord{
					{Name: "example.com", Type: "TXT", Value: "v=spf1 include:mailgun.org ~all", Valid: stringPtr("valid")},
					{Name: "k1._domainkey.example.com", Type: "TXT", Value: "k=rsa; p=...", Valid: stringPtr("valid")},
				},
			},
			expectedError: false,
		},
		{
			name:       "mixed validity returns DNSVerified false",
			domainName: "partial.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"domain": map[string]interface{}{
						"name":  "partial.com",
						"state": "unverified",
					},
					"sending_dns_records": []map[string]interface{}{
						{"name": "partial.com", "record_type": "TXT", "value": "v=spf1", "valid": "valid"},
						{"name": "k1._domainkey.partial.com", "record_type": "TXT", "value": "k=rsa; p=...", "valid": "unknown"},
					},
				})
			},
			expectedDomain: &domaintypes.DomainObservation{
				ID:          "partial.com",
				State:       "unverified",
				DNSVerified: boolPtr(false),
				SendingDNSRecords: []domaintypes.DNSRecord{
					{Name: "partial.com", Type: "TXT", Value: "v=spf1", Valid: stringPtr("valid")},
					{Name: "k1._domainkey.partial.com", Type: "TXT", Value: "k=rsa; p=...", Valid: stringPtr("unknown")},
				},
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
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

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
		domainSpec     *domaintypes.DomainParameters
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedDomain *domaintypes.DomainObservation
		expectedError  bool
	}{
		{
			name:       "successful update returns DNS records",
			domainName: "update.com",
			domainSpec: &domaintypes.DomainParameters{
				SpamAction: stringPtr("tag"),
				WebScheme:  stringPtr("https"),
				Wildcard:   boolPtr(false),
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, "/v4/domains/update.com", r.URL.Path)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

				_ = r.ParseForm()
				assert.Equal(t, "tag", r.FormValue("spam_action"))
				assert.Equal(t, "https", r.FormValue("web_scheme"))
				assert.Equal(t, "false", r.FormValue("wildcard"))

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"domain": map[string]interface{}{
						"name":  "update.com",
						"type":  "sending",
						"state": "active",
					},
					"sending_dns_records": []map[string]interface{}{
						{"name": "update.com", "record_type": "TXT", "value": "v=spf1", "valid": "valid"},
					},
				})
			},
			expectedDomain: &domaintypes.DomainObservation{
				ID:           "update.com",
				State:        "active",
				DNSVerified:  boolPtr(true),
				SendingDNSRecords: []domaintypes.DNSRecord{
					{Name: "update.com", Type: "TXT", Value: "v=spf1", Valid: stringPtr("valid")},
				},
			},
			expectedError: false,
		},
		{
			name:       "domain not found",
			domainName: "notfound.com",
			domainSpec: &domaintypes.DomainParameters{
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
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

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
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Domain has been deleted"})
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
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			err := client.DeleteDomain(context.Background(), tt.domainName)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVerifyDomain(t *testing.T) {
	tests := []struct {
		name                  string
		domainName            string
		serverResponse        func(w http.ResponseWriter, r *http.Request)
		expectedDNSRecCount   int
		expectedDNSVerified   *bool
		expectedError         bool
	}{
		{
			name:       "successful verify returns DNS records with current validity",
			domainName: "verify.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, "/v4/domains/verify.com/verify", r.URL.Path)

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"message": "Domain DNS records have been updated",
					"domain": map[string]interface{}{
						"name":          "verify.com",
						"id":            "verify-id",
						"state":         "unverified",
						"smtp_login":    "postmaster@verify.com",
						"smtp_password": "secret",
					},
					"receiving_dns_records": []map[string]interface{}{
						{"name": "verify.com", "record_type": "MX", "value": "mxa.mailgun.org", "priority": "10", "valid": "unknown"},
					},
					"sending_dns_records": []map[string]interface{}{
						{"name": "verify.com", "record_type": "TXT", "value": "v=spf1", "valid": "unknown"},
					},
				})
			},
			expectedDNSRecCount: 2,
			expectedDNSVerified: boolPtr(false),
			expectedError:       false,
		},
		{
			name:       "verify with all records valid",
			domainName: "allvalid.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"message": "Domain DNS records have been updated",
					"domain": map[string]interface{}{
						"name":  "allvalid.com",
						"state": "active",
					},
					"sending_dns_records": []map[string]interface{}{
						{"name": "allvalid.com", "record_type": "TXT", "value": "v=spf1", "valid": "valid"},
					},
				})
			},
			expectedDNSRecCount: 1,
			expectedDNSVerified: boolPtr(true),
			expectedError:       false,
		},
		{
			name:       "verify endpoint error",
			domainName: "badverify.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal server error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			result, err := client.VerifyDomain(context.Background(), tt.domainName)

			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedDNSRecCount,
				len(result.ReceivingDNSRecords)+len(result.SendingDNSRecords))
			assert.Equal(t, tt.expectedDNSVerified, result.DNSVerified)
		})
	}
}

func TestRecordIsValid(t *testing.T) {
	cases := []struct {
		name string
		in   *string
		want *bool
	}{
		{"nil pointer returns nil", nil, nil},
		{"valid string returns true ptr", stringPtr("valid"), boolPtr(true)},
		{"unknown string returns false ptr", stringPtr("unknown"), boolPtr(false)},
		{"empty string returns false ptr", stringPtr(""), boolPtr(false)},
		{"arbitrary string returns false ptr", stringPtr("foo"), boolPtr(false)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := RecordIsValid(c.in)
			if c.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, *c.want, *got)
		})
	}
}
