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

// Standalone tests that don't depend on the main package
package clients_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Minimal client implementation for testing
type testConfig struct {
	APIKey  string
	BaseURL string
}

type testClient struct {
	config *testConfig
	client *http.Client
}

func newTestClient(config *testConfig) *testClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.mailgun.net/v3"
	}

	return &testClient{
		config: config,
		client: &http.Client{},
	}
}

func (c *testClient) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.config.BaseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("api", c.config.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return c.client.Do(req)
}

// Test basic client functionality without depending on the full codebase
func TestStandaloneClient(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth
		username, password, ok := r.BasicAuth()
		if !ok || username != "api" || password != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Return test response
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "success",
		})
	}))
	defer server.Close()

	// Create client
	client := newTestClient(&testConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	// Test request
	resp, err := client.makeRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestStandaloneClientAuth(t *testing.T) {
	// Test that authentication is properly set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		response := map[string]interface{}{
			"auth_ok":  ok,
			"username": username,
			"password": password,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(&testConfig{
		APIKey:  "my-secret-key",
		BaseURL: server.URL,
	})

	resp, err := client.makeRequest("GET", "/auth-test", nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !result["auth_ok"].(bool) {
		t.Error("Authentication not properly set")
	}

	if result["username"].(string) != "api" {
		t.Errorf("Expected username 'api', got '%s'", result["username"])
	}

	if result["password"].(string) != "my-secret-key" {
		t.Errorf("Expected password 'my-secret-key', got '%s'", result["password"])
	}
}

func TestStandalonePostRequest(t *testing.T) {
	// Test POST requests with form data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/x-www-form-urlencoded" {
			t.Errorf("Expected form content type, got %s", contentType)
		}

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "name=test.com") {
			t.Errorf("Expected form data not found in body: %s", string(body))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"message": "created"}`)
	}))
	defer server.Close()

	client := newTestClient(&testConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	formData := strings.NewReader("name=test.com&type=sending")
	resp, err := client.makeRequest("POST", "/domains", formData)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// Test error handling
func TestStandaloneErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
	}{
		{
			name: "success response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedError: false,
		},
		{
			name: "404 not found",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Not Found"))
			},
			expectedError: true,
		},
		{
			name: "500 server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal Server Error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := newTestClient(&testConfig{
				APIKey:  "test-key",
				BaseURL: server.URL,
			})

			resp, err := client.makeRequest("GET", "/test", nil)
			if err != nil {
				if !tt.expectedError {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}
			defer func() { _ = resp.Body.Close() }()

			// Check if we should expect an error based on status code
			shouldError := resp.StatusCode >= 400
			if shouldError != tt.expectedError {
				t.Errorf("Expected error=%v, got status=%d", tt.expectedError, resp.StatusCode)
			}
		})
	}
}
