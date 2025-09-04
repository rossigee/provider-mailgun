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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestHealthzCheck(t *testing.T) {
	t.Run("HealthzAlwaysSucceeds", func(t *testing.T) {
		checker := NewHealthChecker(nil, nil)

		req := httptest.NewRequest("GET", "/healthz", nil)
		err := checker.HealthzCheck(req)

		assert.NoError(t, err)
	})
}

func TestReadyzCheck(t *testing.T) {
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
		err := checker.ReadyzCheck(req)

		assert.NoError(t, err)
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
		err := checker.ReadyzCheck(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kubernetes unhealthy")
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
		err := checker.ReadyzCheck(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mailgun API unhealthy")
	})

	t.Run("ReadyWithNilChecks", func(t *testing.T) {
		checker := NewHealthChecker(nil, nil)

		req := httptest.NewRequest("GET", "/readyz", nil)
		err := checker.ReadyzCheck(req)

		assert.NoError(t, err)
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
