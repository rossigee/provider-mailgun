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

package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestOperationTimer(t *testing.T) {
	// Reset metrics for clean test
	ResourceOperations.Reset()
	MailgunAPILatency.Reset()

	t.Run("RecordResourceOperation", func(t *testing.T) {
		timer := NewOperationTimer()

		// Sleep briefly to ensure measurable duration
		time.Sleep(1 * time.Millisecond)

		timer.RecordResourceOperation("smtpcredential", "create", "success")

		// Check that counter was incremented
		counter := testutil.ToFloat64(ResourceOperations.WithLabelValues("smtpcredential", "create", "success"))
		assert.Equal(t, float64(1), counter)
	})

	t.Run("RecordMailgunAPIRequest", func(t *testing.T) {
		timer := NewOperationTimer()

		// Sleep briefly to ensure measurable duration
		time.Sleep(1 * time.Millisecond)

		timer.RecordMailgunAPIRequest("create_domain", "example.com", "success")

		// Check that counter was incremented
		counter := testutil.ToFloat64(MailgunAPIRequests.WithLabelValues("create_domain", "example.com", "success"))
		assert.Equal(t, float64(1), counter)
	})

	t.Run("RecordSecretOperation", func(t *testing.T) {
		// Test the static function
		RecordSecretOperation("create", "success")

		// Check that counter was incremented
		counter := testutil.ToFloat64(SecretOperations.WithLabelValues("create", "success"))
		assert.Equal(t, float64(1), counter)
	})

	t.Run("SetProviderConfigUsage", func(t *testing.T) {
		SetProviderConfigUsage("my-config", 5.0)

		// Check that gauge was set
		value := testutil.ToFloat64(ProviderConfigUsage.WithLabelValues("my-config"))
		assert.Equal(t, float64(5.0), value)
	})
}

func TestMetricsRegistration(t *testing.T) {
	t.Run("MetricsAreRegistered", func(t *testing.T) {
		// Check that our metrics are registered with Prometheus
		registry := prometheus.NewRegistry()

		// Register our metrics
		registry.MustRegister(ResourceOperations)
		registry.MustRegister(MailgunAPILatency)
		registry.MustRegister(SecretOperations)
		registry.MustRegister(ProviderConfigUsage)

		// Gather metrics to ensure they're properly registered
		metricFamilies, err := registry.Gather()
		assert.NoError(t, err)
		assert.NotEmpty(t, metricFamilies)

		// Check that we have our expected metrics
		metricNames := make(map[string]bool)
		for _, mf := range metricFamilies {
			metricNames[*mf.Name] = true
		}

		assert.True(t, metricNames["provider_mailgun_resource_operations_total"])
		assert.True(t, metricNames["provider_mailgun_mailgun_api_latency_seconds"])
		assert.True(t, metricNames["provider_mailgun_secret_operations_total"])
		assert.True(t, metricNames["provider_mailgun_provider_config_usage"])
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("RecordResourceOperationFunction", func(t *testing.T) {
		// Reset metrics
		ResourceOperations.Reset()
		OperationDuration.Reset()

		// Record operation
		RecordResourceOperation("domain", "update", "success", 100*time.Millisecond)

		// Check counter
		counter := testutil.ToFloat64(ResourceOperations.WithLabelValues("domain", "update", "success"))
		assert.Equal(t, float64(1), counter)

		// For histograms, we check the count of samples recorded
		// Convert to metric for testing
		metric := &dto.Metric{}
		OperationDuration.WithLabelValues("domain", "update").(prometheus.Histogram).Write(metric)
		assert.Equal(t, uint64(1), *metric.Histogram.SampleCount)
	})
}

func TestConcurrentMetrics(t *testing.T) {
	t.Run("ConcurrentOperations", func(t *testing.T) {
		// Reset metrics
		ResourceOperations.Reset()

		// Run concurrent operations
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				timer := NewOperationTimer()
				timer.RecordResourceOperation("domain", "create", "success")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Check that all operations were recorded
		counter := testutil.ToFloat64(ResourceOperations.WithLabelValues("domain", "create", "success"))
		assert.Equal(t, float64(10), counter)
	})
}
