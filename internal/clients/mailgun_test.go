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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
			expectedError: false, // makeRequest doesn't handle status codes, just returns response
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
			name:           "successful response with JSON",
			statusCode:     200,
			responseBody:   `{"name":"test.com","status":"active"}`,
			target:         &map[string]string{},
			expectedError:  false,
			expectedTarget: &map[string]string{"name": "test.com", "status": "active"},
		},
		{
			name:          "successful response without target",
			statusCode:    204,
			responseBody:  "",
			target:        nil,
			expectedError: false,
		},
		{
			name:          "client error",
			statusCode:    400,
			responseBody:  "Bad Request",
			target:        nil,
			expectedError: true,
		},
		{
			name:          "server error",
			statusCode:    500,
			responseBody:  "Internal Server Error",
			target:        nil,
			expectedError: true,
		},
		{
			name:          "not found error",
			statusCode:    404,
			responseBody:  "Not Found",
			target:        nil,
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

func TestParseRateLimit(t *testing.T) {
	now := time.Now()
	resetEpoch := now.Add(time.Hour).Unix()

	cases := []struct {
		name      string
		status    int
		headers   map[string]string
		wantNil   bool
		wantAfter time.Duration
		wantReset bool
	}{
		{
			name:    "non-429 returns nil",
			status:  http.StatusOK,
			wantNil: true,
		},
		{
			name: "retry-after seconds only",
			status: http.StatusTooManyRequests,
			headers: map[string]string{
				"Retry-After": "30",
			},
			wantAfter: 30 * time.Second,
		},
		{
			name: "x-ratelimit-reset epoch only",
			status: http.StatusTooManyRequests,
			headers: map[string]string{
				"X-RateLimit-Reset": strconv.FormatInt(resetEpoch, 10),
			},
			wantAfter: time.Until(time.Unix(resetEpoch, 0)),
			wantReset: true,
		},
		{
			name: "retry-after wins when both present",
			status: http.StatusTooManyRequests,
			headers: map[string]string{
				"Retry-After":       "5",
				"X-RateLimit-Reset": strconv.FormatInt(resetEpoch, 10),
			},
			wantAfter: 5 * time.Second,
			wantReset: true,
		},
		{
			name: "limit and remaining captured",
			status: http.StatusTooManyRequests,
			headers: map[string]string{
				"Retry-After":         "10",
				"X-RateLimit-Limit":   "100",
				"X-RateLimit-Remaining": "0",
			},
			wantAfter: 10 * time.Second,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: c.status,
				Header:     http.Header{},
			}
			for k, v := range c.headers {
				resp.Header.Set(k, v)
			}

			got := parseRateLimit(resp)
			if c.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil RateLimitError")
			}
			// Allow +/-500ms slop on Reset-derived durations.
			if d := got.RetryAfter - c.wantAfter; d < -500*time.Millisecond || d > 500*time.Millisecond {
				t.Errorf("RetryAfter = %s, want ~%s", got.RetryAfter, c.wantAfter)
			}
			if c.wantReset && got.Reset.IsZero() {
				t.Error("expected Reset to be populated")
			}
			if !c.wantReset && !got.Reset.IsZero() {
				t.Errorf("expected Reset to be zero, got %s", got.Reset)
			}
		})
	}
}

func TestRateLimitError_Error(t *testing.T) {
	e := &RateLimitError{
		RetryAfter: 30 * time.Second,
		Reset:      time.Unix(1700000000, 0),
		Limit:      "100",
		Remaining:  "0",
	}
	msg := e.Error()
	if !strings.Contains(msg, "429") {
		t.Errorf("expected message to contain 429, got %q", msg)
	}
	if !strings.Contains(msg, "retry after 30s") {
		t.Errorf("expected message to contain retry-after, got %q", msg)
	}
	if !strings.Contains(msg, "limit=100 remaining=0") {
		t.Errorf("expected message to contain limit/remaining, got %q", msg)
	}
}

func TestMakeRequestRateLimitRetry(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1") // 1 second is enough to exercise the back-off path
			w.Header().Set("X-RateLimit-Limit", "100")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limit"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	}
	c := NewClient(config).(*mailgunClient)

	start := time.Now()
	resp, err := c.makeRequest(context.Background(), "GET", "/test", nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	_ = resp.Body.Close()
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
	// Retry-After=1 should produce roughly 1s elapsed. Allow 200ms slop for slow CI.
	if elapsed < 800*time.Millisecond {
		t.Errorf("expected retry to wait ~1s, elapsed=%s", elapsed)
	}
}

func TestMakeRequestRateLimitExhaustion(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	config := &Config{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	}
	c := NewClient(config).(*mailgunClient)

	resp, err := c.makeRequest(context.Background(), "GET", "/test", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	// Initial attempt + 3 retries = 4 calls.
	if got := atomic.LoadInt32(&calls); got != 4 {
		t.Errorf("expected 4 calls (initial + 3 retries), got %d", got)
	}
}

// testError is a helper for testing error conditions
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestRegions_HasUSAndEU(t *testing.T) {
	codes := make(map[string]bool, len(regions))
	for _, r := range regions {
		if r.Code == "" {
			t.Errorf("region entry has empty Code: %+v", r)
		}
		if r.BaseURL == "" || r.V4BaseURL == "" || r.SMTPHost == "" {
			t.Errorf("region %q missing required field: %+v", r.Code, r)
		}
		codes[r.Code] = true
	}
	for _, want := range []string{"US", "EU"} {
		if !codes[want] {
			t.Errorf("regions registry missing %q; got %v", want, codes)
		}
	}
	if regions[0].SMTPHost != DefaultSMTPHost {
		t.Errorf("first region SMTPHost %q does not match DefaultSMTPHost %q — keep them in sync",
			regions[0].SMTPHost, DefaultSMTPHost)
	}
}

func TestFindRegionByCode(t *testing.T) {
	tests := []struct {
		code     string
		wantCode string
		wantOK   bool
	}{
		{code: "US", wantCode: "US", wantOK: true},
		{code: "EU", wantCode: "EU", wantOK: true},
		{code: "us", wantOK: false},
		{code: "", wantOK: false},
		{code: "AP", wantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			got, ok := findRegionByCode(tc.code)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && got.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", got.Code, tc.wantCode)
			}
		})
	}
}

func TestDetectRegionByURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		wantCode string
		wantOK   bool
	}{
		{name: "US default", baseURL: "https://api.mailgun.net/v3", wantCode: "US", wantOK: true},
		{name: "US v4", baseURL: "https://api.mailgun.net/v4", wantCode: "US", wantOK: true},
		{name: "EU v3", baseURL: "https://api.eu.mailgun.net/v3", wantCode: "EU", wantOK: true},
		{name: "EU v4", baseURL: "https://api.eu.mailgun.net/v4", wantCode: "EU", wantOK: true},
		{name: "EU marker wins over US marker", baseURL: "https://api.eu.mailgun.net/v3", wantCode: "EU", wantOK: true},
		{name: "custom URL no match", baseURL: "https://internal.example.com/v3", wantOK: false},
		{name: "empty", baseURL: "", wantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := detectRegionByURL(tc.baseURL)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v (got %+v)", ok, tc.wantOK, got)
			}
			if ok && got.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", got.Code, tc.wantCode)
			}
		})
	}
}

func TestDeriveV4BaseURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "v3 suffix swapped", in: "https://api.mailgun.net/v3", want: "https://api.mailgun.net/v4"},
		{name: "EU v3 swapped", in: "https://api.eu.mailgun.net/v3", want: "https://api.eu.mailgun.net/v4"},
		{name: "v4 kept verbatim", in: "https://api.mailgun.net/v4", want: "https://api.mailgun.net/v4"},
		{name: "no suffix appended", in: "https://api.mailgun.net", want: "https://api.mailgun.net/v4"},
		{name: "trailing path no suffix", in: "https://api.mailgun.net/custom", want: "https://api.mailgun.net/custom/v4"},
		{name: "empty", in: "", want: "/v4"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DeriveV4BaseURL(tc.in)
			if got != tc.want {
				t.Errorf("DeriveV4BaseURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
