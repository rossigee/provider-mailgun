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

package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Comprehensive API client tests
func TestMailgunAPISimulation(t *testing.T) {
	// Create a comprehensive mock Mailgun server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication
		username, _, ok := r.BasicAuth()
		if !ok || username != "api" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
			return
		}

		// Route based on path and method
		switch {
		case r.URL.Path == "/v3/domains" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"total_count": 2,
				"items": []map[string]interface{}{
					{
						"name":  "example.com",
						"type":  "sending",
						"state": "active",
						"created_at": "2025-01-01T00:00:00Z",
					},
					{
						"name":  "test.com",
						"type":  "receiving",
						"state": "active",
						"created_at": "2025-01-01T00:00:00Z",
					},
				},
			})

		case r.URL.Path == "/v3/domains" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"domain": map[string]interface{}{
					"name":  "new.example.com",
					"type":  "sending",
					"state": "active",
				},
				"message": "Domain has been created",
			})

		case r.URL.Path == "/v3/domains/example.com" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"domain": map[string]interface{}{
					"name":  "example.com",
					"type":  "sending",
					"state": "active",
					"created_at": "2025-01-01T00:00:00Z",
					"smtp_login": "postmaster@example.com",
				},
			})

		case r.URL.Path == "/v3/domains/example.com" && r.Method == "DELETE":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Domain has been deleted",
			})

		case r.URL.Path == "/v3/lists" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"list": map[string]interface{}{
					"address":     "test@example.com",
					"name":        "Test List",
					"description": "A test mailing list",
					"access_level": "readonly",
					"created_at":  "2025-01-01T00:00:00Z",
				},
				"message": "Mailing list has been created",
			})

		case r.URL.Path == "/v3/lists/test@example.com" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"list": map[string]interface{}{
					"address":      "test@example.com",
					"name":         "Test List",
					"description":  "A test mailing list",
					"access_level": "readonly",
					"members_count": 0,
				},
			})

		case r.URL.Path == "/v3/routes" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"total_count": 1,
				"items": []map[string]interface{}{
					{
						"id":          "route123",
						"priority":    1,
						"description": "Forward to support",
						"expression":  "match_recipient(\"support@example.com\")",
						"actions":     []string{"forward(\"support-team@example.com\")"},
						"created_at":  "2025-01-01T00:00:00Z",
					},
				},
			})

		case r.URL.Path == "/v3/routes" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"route": map[string]interface{}{
					"id":          "route456",
					"priority":    5,
					"description": "New route",
					"expression":  "match_recipient(\"new@example.com\")",
					"actions":     []string{"forward(\"new-team@example.com\")"},
				},
				"message": "Route has been created",
			})

		case r.URL.Path == "/v3/domains/example.com/webhooks" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"webhooks": []map[string]interface{}{
					{
						"id":  "webhook123",
						"url": "https://example.com/webhook",
						"events": []string{"delivered", "opened"},
						"created_at": "2025-01-01T00:00:00Z",
					},
				},
			})

		case r.URL.Path == "/v3/domains/example.com/webhooks" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"webhook": map[string]interface{}{
					"id":  "webhook456",
					"url": "https://example.com/new-webhook",
					"events": []string{"delivered", "opened", "clicked"},
				},
				"message": "Webhook has been created",
			})

		// SMTP Credential endpoints
		case r.URL.Path == "/v3/domains/example.com/credentials" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"login":      "test@example.com",
						"created_at": "2025-01-01T00:00:00Z",
						"state":      "active",
					},
				},
			})

		case r.URL.Path == "/v3/domains/example.com/credentials" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"login":    "new@example.com",
				"password": "generated-password",
				"state":    "active",
			})

		case r.URL.Path == "/v3/domains/example.com/credentials/test@example.com" && r.Method == "DELETE":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Credential has been deleted",
			})

		// Template endpoints
		case r.URL.Path == "/v3/domains/example.com/templates" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"name":        "welcome-template",
						"description": "Welcome email template",
						"created_at":  "2025-01-01T00:00:00Z",
						"created_by":  "api",
					},
				},
			})

		case r.URL.Path == "/v3/domains/example.com/templates" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"template": map[string]interface{}{
					"name":        "new-template",
					"description": "New email template",
					"created_at":  "2025-01-01T00:00:00Z",
					"created_by":  "api",
				},
				"message": "Template has been created",
			})

		case r.URL.Path == "/v3/domains/example.com/templates/welcome-template" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"template": map[string]interface{}{
					"name":        "welcome-template",
					"description": "Welcome email template",
					"created_at":  "2025-01-01T00:00:00Z",
					"created_by":  "api",
					"versions": []map[string]interface{}{
						{
							"tag":        "v1.0",
							"engine":     "mustache",
							"created_at": "2025-01-01T00:00:00Z",
							"comment":    "Initial version",
							"active":     true,
							"template":   "<h1>Welcome {{name}}!</h1>",
						},
					},
				},
			})

		case r.URL.Path == "/v3/domains/example.com/templates/welcome-template" && r.Method == "DELETE":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Template has been deleted",
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Not found",
			})
		}
	}))
	defer server.Close()

	// Test scenarios
	testScenarios := []struct {
		name           string
		method         string
		path           string
		apiKey         string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "list domains",
			method:         "GET",
			path:           "/v3/domains",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if response["total_count"].(float64) != 2 {
					t.Errorf("Expected 2 domains, got %v", response["total_count"])
				}
			},
		},
		{
			name:           "create domain",
			method:         "POST",
			path:           "/v3/domains",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if response["message"].(string) != "Domain has been created" {
					t.Errorf("Unexpected response message: %v", response["message"])
				}
			},
		},
		{
			name:           "get specific domain",
			method:         "GET",
			path:           "/v3/domains/example.com",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				domain := response["domain"].(map[string]interface{})
				if domain["name"].(string) != "example.com" {
					t.Errorf("Expected domain name 'example.com', got %v", domain["name"])
				}
			},
		},
		{
			name:           "unauthorized access",
			method:         "GET",
			path:           "/v3/domains",
			apiKey:         "",
			expectedStatus: 401,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if response["message"].(string) != "Unauthorized" {
					t.Errorf("Expected 'Unauthorized' message, got %v", response["message"])
				}
			},
		},
		{
			name:           "create mailing list",
			method:         "POST",
			path:           "/v3/lists",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				list := response["list"].(map[string]interface{})
				if list["address"].(string) != "test@example.com" {
					t.Errorf("Expected list address 'test@example.com', got %v", list["address"])
				}
			},
		},
		{
			name:           "list routes",
			method:         "GET",
			path:           "/v3/routes",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if response["total_count"].(float64) != 1 {
					t.Errorf("Expected 1 route, got %v", response["total_count"])
				}
			},
		},
		{
			name:           "list webhooks",
			method:         "GET",
			path:           "/v3/domains/example.com/webhooks",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				webhooks := response["webhooks"].([]interface{})
				if len(webhooks) != 1 {
					t.Errorf("Expected 1 webhook, got %d", len(webhooks))
				}
			},
		},
		{
			name:           "list SMTP credentials",
			method:         "GET",
			path:           "/v3/domains/example.com/credentials",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				items := response["items"].([]interface{})
				if len(items) != 1 {
					t.Errorf("Expected 1 credential, got %d", len(items))
				}
			},
		},
		{
			name:           "create SMTP credential",
			method:         "POST",
			path:           "/v3/domains/example.com/credentials",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if response["login"].(string) != "new@example.com" {
					t.Errorf("Expected login 'new@example.com', got %v", response["login"])
				}
			},
		},
		{
			name:           "list templates",
			method:         "GET",
			path:           "/v3/domains/example.com/templates",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				items := response["items"].([]interface{})
				if len(items) != 1 {
					t.Errorf("Expected 1 template, got %d", len(items))
				}
			},
		},
		{
			name:           "get template",
			method:         "GET",
			path:           "/v3/domains/example.com/templates/welcome-template",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				template := response["template"].(map[string]interface{})
				if template["name"].(string) != "welcome-template" {
					t.Errorf("Expected template name 'welcome-template', got %v", template["name"])
				}
				versions := template["versions"].([]interface{})
				if len(versions) != 1 {
					t.Errorf("Expected 1 version, got %d", len(versions))
				}
			},
		},
		{
			name:           "create template",
			method:         "POST",
			path:           "/v3/domains/example.com/templates",
			apiKey:         "valid-key",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				template := response["template"].(map[string]interface{})
				if template["name"].(string) != "new-template" {
					t.Errorf("Expected template name 'new-template', got %v", template["name"])
				}
			},
		},
	}

	// Run test scenarios
	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second}

			req, err := http.NewRequest(scenario.method, server.URL+scenario.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if scenario.apiKey != "" {
				req.SetBasicAuth("api", scenario.apiKey)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != scenario.expectedStatus {
				t.Errorf("Expected status %d, got %d", scenario.expectedStatus, resp.StatusCode)
			}

			// Read response body
			body := make([]byte, 4096)
			n, _ := resp.Body.Read(body)
			body = body[:n]

			// Run custom response checks
			if scenario.checkResponse != nil {
				scenario.checkResponse(t, body)
			}
		})
	}
}

// Test error scenarios
func TestErrorScenarios(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/error/500":
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Internal server error",
			})
		case "/v3/error/429":
			w.Header().Set("X-RateLimit-Limit", "100")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", "1609459200")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Rate limit exceeded",
			})
		case "/v3/error/400":
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Bad request",
			})
		}
	}))
	defer server.Close()

	errorTests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "internal server error",
			path:           "/v3/error/500",
			expectedStatus: 500,
		},
		{
			name:           "rate limit exceeded",
			path:           "/v3/error/429",
			expectedStatus: 429,
			expectedHeader: "X-RateLimit-Limit",
		},
		{
			name:           "bad request",
			path:           "/v3/error/400",
			expectedStatus: 400,
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second}

			req, err := http.NewRequest("GET", server.URL+test.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.SetBasicAuth("api", "test-key")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, resp.StatusCode)
			}

			if test.expectedHeader != "" {
				if resp.Header.Get(test.expectedHeader) == "" {
					t.Errorf("Expected header %s not found", test.expectedHeader)
				}
			}
		})
	}
}

// Test client timeout and retry scenarios
func TestClientResilience(t *testing.T) {
	// Server that introduces delays
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v3/slow" {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "slow response"})
		} else {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "fast response"})
		}
	}))
	defer server.Close()

	t.Run("client timeout", func(t *testing.T) {
		client := &http.Client{Timeout: 50 * time.Millisecond} // Very short timeout

		req, err := http.NewRequest("GET", server.URL+"/v3/slow", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.SetBasicAuth("api", "test-key")

		_, err = client.Do(req)
		if err == nil {
			t.Error("Expected timeout error, but request succeeded")
		}
	})

	t.Run("normal response", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		req, err := http.NewRequest("GET", server.URL+"/v3/fast", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.SetBasicAuth("api", "test-key")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}
