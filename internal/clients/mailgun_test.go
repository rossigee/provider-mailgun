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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "404 error",
			err:      &testError{msg: "API request failed with status 404: Not Found"},
			expected: true,
		},
		{
			name:     "other error",
			err:      &testError{msg: "API request failed with status 500: Internal Server Error"},
			expected: false,
		},
		{
			name:     "not found in message",
			err:      &testError{msg: "resource not found"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	config := &Config{
		APIKey:  "test-key",
		BaseURL: "https://api.mailgun.net/v3",
	}

	client := NewClient(config)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	mgClient, ok := client.(*mailgunClient)
	if !ok {
		t.Fatal("NewClient() did not return a mailgunClient")
	}

	if mgClient.config.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", mgClient.config.APIKey)
	}

	if mgClient.config.BaseURL != "https://api.mailgun.net/v3" {
		t.Errorf("Expected base URL 'https://api.mailgun.net/v3', got '%s'", mgClient.config.BaseURL)
	}

	if mgClient.config.HTTPClient == nil {
		t.Error("Expected HTTP client to be set")
	}
}

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           io.Reader
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
	}{
		{
			name:   "successful GET request",
			method: "GET",
			path:   "/domains",
			body:   nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/v3/domains" {
					t.Errorf("Expected path '/v3/domains', got %s", r.URL.Path)
				}

				username, password, ok := r.BasicAuth()
				if !ok || username != "api" || password != "test-key" {
					t.Error("Expected basic auth with username 'api' and password 'test-key'")
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			},
			expectedError: false,
		},
		{
			name:   "successful POST request with body",
			method: "POST",
			path:   "/domains",
			body:   strings.NewReader("name=test.com"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				contentType := r.Header.Get("Content-Type")
				if contentType != "application/x-www-form-urlencoded" {
					t.Errorf("Expected content type 'application/x-www-form-urlencoded', got %s", contentType)
				}

				body, _ := io.ReadAll(r.Body)
				if string(body) != "name=test.com" {
					t.Errorf("Expected body 'name=test.com', got %s", string(body))
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "created"})
			},
			expectedError: false,
		},
		{
			name:   "server error",
			method: "GET",
			path:   "/domains",
			body:   nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal Server Error"))
			},
			expectedError: false,  // makeRequest doesn't handle status codes, just returns response
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create client with test server URL
			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config).(*mailgunClient)

			// Make request
			resp, err := client.makeRequest(context.Background(), tt.method, tt.path, tt.body)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if resp != nil {
				_ = resp.Body.Close()
			}
		})
	}
}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		target         interface{}
		expectedError  bool
		expectedTarget interface{}
	}{
		{
			name:         "successful response with JSON",
			statusCode:   200,
			responseBody: `{"name":"test.com","status":"active"}`,
			target:       &map[string]string{},
			expectedError: false,
			expectedTarget: &map[string]string{"name": "test.com", "status": "active"},
		},
		{
			name:         "successful response without target",
			statusCode:   204,
			responseBody: "",
			target:       nil,
			expectedError: false,
		},
		{
			name:         "client error",
			statusCode:   400,
			responseBody: "Bad Request",
			target:       nil,
			expectedError: true,
		},
		{
			name:         "server error",
			statusCode:   500,
			responseBody: "Internal Server Error",
			target:       nil,
			expectedError: true,
		},
		{
			name:         "not found error",
			statusCode:   404,
			responseBody: "Not Found",
			target:       nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client and make request
			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			}
			client := NewClient(config).(*mailgunClient)

			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			// Test handleResponse
			err = client.handleResponse(resp, tt.target)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectedTarget != nil && tt.target != nil {
				expected := tt.expectedTarget.(*map[string]string)
				actual := tt.target.(*map[string]string)

				for k, v := range *expected {
					if (*actual)[k] != v {
						t.Errorf("Expected %s=%s, got %s=%s", k, v, k, (*actual)[k])
					}
				}
			}
		})
	}
}

func TestCreateFormData(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name: "simple string values",
			params: map[string]interface{}{
				"name": "test.com",
				"type": "sending",
			},
			expected: "name=test.com&type=sending",
		},
		{
			name: "mixed types",
			params: map[string]interface{}{
				"name":     "test.com",
				"priority": 10,
				"enabled":  true,
			},
			expected: "enabled=true&name=test.com&priority=10",
		},
		{
			name: "nil values ignored",
			params: map[string]interface{}{
				"name":        "test.com",
				"description": nil,
				"type":        "sending",
			},
			expected: "name=test.com&type=sending",
		},
		{
			name:     "empty params",
			params:   map[string]interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createFormData(tt.params)

			// Since maps are unordered, we need to check that all expected
			// key-value pairs are present in the result
			if tt.expected == "" {
				if result != "" {
					t.Errorf("Expected empty string, got '%s'", result)
				}
				return
			}

			// Parse expected and actual to compare
			expectedPairs := strings.Split(tt.expected, "&")
			actualPairs := strings.Split(result, "&")

			if len(expectedPairs) != len(actualPairs) {
				t.Errorf("Expected %d pairs, got %d pairs", len(expectedPairs), len(actualPairs))
				return
			}

			// Convert to maps for easier comparison
			expectedMap := make(map[string]string)
			for _, pair := range expectedPairs {
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					expectedMap[parts[0]] = parts[1]
				}
			}

			actualMap := make(map[string]string)
			for _, pair := range actualPairs {
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					actualMap[parts[0]] = parts[1]
				}
			}

			for k, v := range expectedMap {
				if actualMap[k] != v {
					t.Errorf("Expected %s=%s, got %s=%s", k, v, k, actualMap[k])
				}
			}
		})
	}
}

// testError is a helper for testing error conditions
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
