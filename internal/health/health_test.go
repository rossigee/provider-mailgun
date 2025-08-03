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

package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rossigee/provider-mailgun/internal/clients"
)

// mockKubeClient implements the client.Client interface for testing
type mockKubeClient struct {
	client.Client
	restMapperFunc func() (map[schema.GroupKind][]schema.GroupVersionResource, error)
}

func (m *mockKubeClient) RESTMapper() meta.RESTMapper {
	return &mockRESTMapper{mappingFunc: m.restMapperFunc}
}

type mockRESTMapper struct {
	meta.RESTMapper
	mappingFunc func() (map[schema.GroupKind][]schema.GroupVersionResource, error)
}

func (m *mockRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	if m.mappingFunc != nil {
		_, err := m.mappingFunc()
		if err != nil {
			return nil, err
		}
	}
	// Return a dummy mapping for successful cases
	return []*meta.RESTMapping{
		{
			Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"},
		},
	}, nil
}

func TestHealthChecker(t *testing.T) {
	t.Run("NewHealthChecker", func(t *testing.T) {
		mockClient := &mockKubeClient{}
		mailgunCheck := func(ctx context.Context) error { return nil }

		checker := NewHealthChecker(mockClient, mailgunCheck)

		assert.NotNil(t, checker)
		assert.Equal(t, mockClient, checker.kubeClient)
		assert.NotNil(t, checker.mailgunCheck)
	})
}

func TestServeHealthz(t *testing.T) {
	t.Run("HealthzEndpoint", func(t *testing.T) {
		checker := NewHealthChecker(nil, nil)

		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()

		checker.ServeHealthz(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var status HealthStatus
		err := json.Unmarshal(w.Body.Bytes(), &status)
		require.NoError(t, err)

		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "Provider is running", status.Message)
		assert.NotZero(t, status.Timestamp)
		assert.NotEmpty(t, status.Duration)
	})
}

func TestServeReadyz(t *testing.T) {
	t.Run("ReadyWithHealthyComponents", func(t *testing.T) {
		mockClient := &mockKubeClient{
			restMapperFunc: func() (map[schema.GroupKind][]schema.GroupVersionResource, error) {
				return nil, nil // Success
			},
		}

		mailgunCheck := func(ctx context.Context) error {
			return nil // Success
		}

		checker := NewHealthChecker(mockClient, mailgunCheck)

		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		checker.ServeReadyz(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var status HealthStatus
		err := json.Unmarshal(w.Body.Bytes(), &status)
		require.NoError(t, err)

		assert.Equal(t, "ready", status.Status)
		assert.Equal(t, "All components are healthy", status.Message)
		assert.Equal(t, "healthy", status.Details["kubernetes"])
		assert.Equal(t, "healthy", status.Details["mailgun_api"])
	})

	t.Run("NotReadyWithUnhealthyKubernetes", func(t *testing.T) {
		mockClient := &mockKubeClient{
			restMapperFunc: func() (map[schema.GroupKind][]schema.GroupVersionResource, error) {
				return nil, fmt.Errorf("kubernetes API unavailable")
			},
		}

		mailgunCheck := func(ctx context.Context) error {
			return nil // Success
		}

		checker := NewHealthChecker(mockClient, mailgunCheck)

		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		checker.ServeReadyz(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var status HealthStatus
		err := json.Unmarshal(w.Body.Bytes(), &status)
		require.NoError(t, err)

		assert.Equal(t, "not_ready", status.Status)
		assert.Equal(t, "Some components are unhealthy", status.Message)
		assert.Contains(t, status.Details["kubernetes"], "unhealthy")
		assert.Equal(t, "healthy", status.Details["mailgun_api"])
	})

	t.Run("NotReadyWithUnhealthyMailgun", func(t *testing.T) {
		mockClient := &mockKubeClient{
			restMapperFunc: func() (map[schema.GroupKind][]schema.GroupVersionResource, error) {
				return nil, nil // Success
			},
		}

		mailgunCheck := func(ctx context.Context) error {
			return fmt.Errorf("mailgun API unreachable")
		}

		checker := NewHealthChecker(mockClient, mailgunCheck)

		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		checker.ServeReadyz(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var status HealthStatus
		err := json.Unmarshal(w.Body.Bytes(), &status)
		require.NoError(t, err)

		assert.Equal(t, "not_ready", status.Status)
		assert.Contains(t, status.Details["mailgun_api"], "unhealthy")
		assert.Equal(t, "healthy", status.Details["kubernetes"])
	})

	t.Run("ReadyWithNilChecks", func(t *testing.T) {
		checker := NewHealthChecker(nil, nil)

		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		checker.ServeReadyz(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var status HealthStatus
		err := json.Unmarshal(w.Body.Bytes(), &status)
		require.NoError(t, err)

		assert.Equal(t, "ready", status.Status)
		assert.Empty(t, status.Details) // No checks performed
	})

	t.Run("ReadinessTimeout", func(t *testing.T) {
		slowMailgunCheck := func(ctx context.Context) error {
			// Sleep longer than the 5-second timeout
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(6 * time.Second):
				return nil
			}
		}

		checker := NewHealthChecker(nil, slowMailgunCheck)

		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		checker.ServeReadyz(w, req)
		duration := time.Since(start)

		// Should complete within reasonable time due to timeout
		assert.Less(t, duration, 7*time.Second)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestCreateMailgunHealthCheck(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		checkFunc := CreateMailgunHealthCheck(nil)
		assert.Nil(t, checkFunc)
	})

	t.Run("ValidConfig", func(t *testing.T) {
		config := &clients.Config{
			APIKey:  "test-key",
			BaseURL: "https://api.mailgun.net/v3",
		}

		checkFunc := CreateMailgunHealthCheck(config)
		assert.NotNil(t, checkFunc)
	})

	t.Run("HealthCheckExecution", func(t *testing.T) {
		// Create a test server that responds with 404 for health checks
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v3/domains/health-check-non-existent-domain.test" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message": "Domain not found"}`))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		config := &clients.Config{
			APIKey:  "test-key",
			BaseURL: server.URL + "/v3",
		}

		checkFunc := CreateMailgunHealthCheck(config)
		require.NotNil(t, checkFunc)

		ctx := context.Background()
		err := checkFunc(ctx)

		// Should succeed because 404 means API is accessible
		assert.NoError(t, err)
	})

	t.Run("HealthCheckFailure", func(t *testing.T) {
		// Create a test server that simulates network failure
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Close connection to simulate network error
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				_ = conn.Close()
			}
		}))
		server.Close() // Close immediately to simulate connection failure

		config := &clients.Config{
			APIKey:  "test-key",
			BaseURL: server.URL + "/v3",
		}

		checkFunc := CreateMailgunHealthCheck(config)
		require.NotNil(t, checkFunc)

		ctx := context.Background()
		err := checkFunc(ctx)

		// Should fail because of connection error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mailgun API not accessible")
	})
}

func TestSetupHealthChecks(t *testing.T) {
	t.Run("SetupEndpoints", func(t *testing.T) {
		mux := http.NewServeMux()
		checker := NewHealthChecker(nil, nil)

		SetupHealthChecks(mux, checker)

		// Test healthz endpoint
		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Test readyz endpoint
		req = httptest.NewRequest("GET", "/readyz", nil)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthStatusJSON(t *testing.T) {
	t.Run("JSONSerialization", func(t *testing.T) {
		status := HealthStatus{
			Status:    "healthy",
			Message:   "All good",
			Timestamp: time.Now(),
			Details: map[string]string{
				"component1": "healthy",
				"component2": "degraded",
			},
			Duration: "5ms",
		}

		data, err := json.Marshal(status)
		require.NoError(t, err)

		var unmarshaled HealthStatus
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, status.Status, unmarshaled.Status)
		assert.Equal(t, status.Message, unmarshaled.Message)
		assert.Equal(t, status.Details, unmarshaled.Details)
		assert.Equal(t, status.Duration, unmarshaled.Duration)
	})
}
