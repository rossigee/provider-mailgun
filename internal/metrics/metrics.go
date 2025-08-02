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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// MetricsNamespace is the namespace for all provider metrics
	MetricsNamespace = "provider_mailgun"

	// Label names
	LabelResource  = "resource"
	LabelOperation = "operation"
	LabelDomain    = "domain"
	LabelProvider  = "provider_config"
	LabelResult    = "result"
)

var (
	// ResourceOperations tracks the total number of resource operations
	ResourceOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "resource_operations_total",
			Help:      "Total number of resource operations performed",
		},
		[]string{LabelResource, LabelOperation, LabelResult},
	)

	// OperationDuration tracks the duration of resource operations
	OperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: MetricsNamespace,
			Name:      "operation_duration_seconds",
			Help:      "Duration of resource operations in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{LabelResource, LabelOperation},
	)

	// MailgunAPIRequests tracks API requests to Mailgun
	MailgunAPIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "mailgun_api_requests_total",
			Help:      "Total number of Mailgun API requests",
		},
		[]string{LabelOperation, LabelDomain, LabelResult},
	)

	// MailgunAPILatency tracks API request latency
	MailgunAPILatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: MetricsNamespace,
			Name:      "mailgun_api_latency_seconds",
			Help:      "Latency of Mailgun API requests in seconds",
			Buckets:   []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{LabelOperation, LabelDomain},
	)

	// SecretOperations tracks secret creation/retrieval for SMTP credentials
	SecretOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "secret_operations_total",
			Help:      "Total number of Kubernetes secret operations",
		},
		[]string{LabelOperation, LabelResult},
	)

	// ProviderConfigUsage tracks ProviderConfig usage
	ProviderConfigUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Name:      "provider_config_usage",
			Help:      "Number of resources using each ProviderConfig",
		},
		[]string{LabelProvider},
	)
)

func init() {
	// Register metrics with controller-runtime metrics registry
	metrics.Registry.MustRegister(
		ResourceOperations,
		OperationDuration,
		MailgunAPIRequests,
		MailgunAPILatency,
		SecretOperations,
		ProviderConfigUsage,
	)
}

// RecordResourceOperation records a resource operation with timing
func RecordResourceOperation(resource, operation, result string, duration time.Duration) {
	ResourceOperations.WithLabelValues(resource, operation, result).Inc()
	OperationDuration.WithLabelValues(resource, operation).Observe(duration.Seconds())
}

// RecordMailgunAPIRequest records a Mailgun API request
func RecordMailgunAPIRequest(operation, domain, result string, duration time.Duration) {
	MailgunAPIRequests.WithLabelValues(operation, domain, result).Inc()
	MailgunAPILatency.WithLabelValues(operation, domain).Observe(duration.Seconds())
}

// RecordSecretOperation records a Kubernetes secret operation
func RecordSecretOperation(operation, result string) {
	SecretOperations.WithLabelValues(operation, result).Inc()
}

// SetProviderConfigUsage sets the usage count for a ProviderConfig
func SetProviderConfigUsage(providerConfig string, count float64) {
	ProviderConfigUsage.WithLabelValues(providerConfig).Set(count)
}

// OperationTimer provides a convenient way to time operations
type OperationTimer struct {
	start time.Time
}

// NewOperationTimer creates a new operation timer
func NewOperationTimer() *OperationTimer {
	return &OperationTimer{start: time.Now()}
}

// RecordResourceOperation records the operation with timing since creation
func (t *OperationTimer) RecordResourceOperation(resource, operation, result string) {
	RecordResourceOperation(resource, operation, result, time.Since(t.start))
}

// RecordMailgunAPIRequest records the API request with timing since creation
func (t *OperationTimer) RecordMailgunAPIRequest(operation, domain, result string) {
	RecordMailgunAPIRequest(operation, domain, result, time.Since(t.start))
}
