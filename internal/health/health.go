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
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rossigee/provider-mailgun/internal/clients"
)



// HealthChecker provides health checking functionality
type HealthChecker struct {
	kubeClient   client.Client
	mailgunCheck func(context.Context) error
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(kubeClient client.Client, mailgunCheckFunc func(context.Context) error) *HealthChecker {
	return &HealthChecker{
		kubeClient:   kubeClient,
		mailgunCheck: mailgunCheckFunc,
	}
}


// checkKubernetes verifies Kubernetes API connectivity
func (h *HealthChecker) checkKubernetes(ctx context.Context) error {
	// Try to get API resources as a basic connectivity test
	_, err := h.kubeClient.RESTMapper().RESTMappings(schema.GroupKind{Group: "", Kind: "Namespace"})
	if err != nil {
		return fmt.Errorf("kubernetes API not accessible: %w", err)
	}
	return nil
}

// CreateMailgunHealthCheck creates a health check function for Mailgun API
func CreateMailgunHealthCheck(config *clients.Config) func(context.Context) error {
	if config == nil {
		return nil
	}

	return func(ctx context.Context) error {
		// Create a client with the provided config
		client := clients.NewClient(config)

		// Try to get a non-existent domain to test API connectivity
		// This should return a 404, which means the API is accessible
		_, err := client.GetDomain(ctx, "health-check-non-existent-domain.test")
		if err != nil {
			// Check if it's a "not found" error, which means API is working
			if clients.IsNotFound(err) {
				return nil // API is accessible
			}
			return fmt.Errorf("mailgun API not accessible: %w", err)
		}

		// If no error, API is accessible (though this is unlikely for a non-existent domain)
		return nil
	}
}

// HealthzCheck returns a healthz.Checker for liveness probes
func (h *HealthChecker) HealthzCheck(req *http.Request) error {
	// Simple liveness check - just ensure the process is running
	return nil
}

// ReadyzCheck returns a healthz.Checker for readiness probes
func (h *HealthChecker) ReadyzCheck(req *http.Request) error {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	// Check Kubernetes connectivity
	if h.kubeClient != nil {
		if err := h.checkKubernetes(ctx); err != nil {
			return fmt.Errorf("kubernetes unhealthy: %w", err)
		}
	}

	// Check Mailgun API connectivity (if available)
	if h.mailgunCheck != nil {
		if err := h.mailgunCheck(ctx); err != nil {
			return fmt.Errorf("mailgun API unhealthy: %w", err)
		}
	}

	return nil
}
