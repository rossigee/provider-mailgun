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

package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/rossigee/provider-mailgun/internal/health"
)

func TestHealthChecker(t *testing.T) {
	// Create a fake Kubernetes client with proper scheme
	s := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(s))
	require.NoError(t, corev1.AddToScheme(s))

	kubeClient := fake.NewClientBuilder().
		WithScheme(s).
		Build()

	// Create health checker
	healthChecker := health.NewHealthChecker(kubeClient, nil)

	t.Run("HealthzCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		err := healthChecker.HealthzCheck(req)

		// Healthz should always succeed (liveness check)
		assert.NoError(t, err)
	})

	t.Run("ReadyzCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/readyz", nil)
		err := healthChecker.ReadyzCheck(req)

		// The fake client doesn't support RESTMapper queries properly,
		// so we expect this to fail with kubernetes unhealthy
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kubernetes unhealthy")
	})
}

func TestHealthCheckerWithMailgunCheck(t *testing.T) {
	// Create a fake Kubernetes client with proper scheme
	s := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(s))
	require.NoError(t, corev1.AddToScheme(s))

	kubeClient := fake.NewClientBuilder().
		WithScheme(s).
		Build()

	// Create a mock Mailgun health check that always succeeds
	mailgunCheck := func(ctx context.Context) error {
		return nil
	}

	// Create health checker with Mailgun check
	healthChecker := health.NewHealthChecker(kubeClient, mailgunCheck)

	t.Run("ReadyzWithMailgunCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/readyz", nil)
		err := healthChecker.ReadyzCheck(req)

		// Even with a healthy Mailgun check, kubernetes check fails with fake client
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kubernetes unhealthy")
	})
}

func TestGetWatchNamespace(t *testing.T) {
	t.Run("NoEnvironmentVariable", func(t *testing.T) {
		// Ensure the environment variable is not set
		t.Setenv("WATCH_NAMESPACE", "")

		ns, err := getWatchNamespace()

		assert.NoError(t, err)
		assert.Empty(t, ns)
	})

	t.Run("WithEnvironmentVariable", func(t *testing.T) {
		expectedNS := "test-namespace"
		t.Setenv("WATCH_NAMESPACE", expectedNS)

		ns, err := getWatchNamespace()

		assert.NoError(t, err)
		assert.Equal(t, expectedNS, ns)
	})
}
