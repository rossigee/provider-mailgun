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

package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.False(t, config.Enabled, "Tracing should be disabled by default")
	assert.Equal(t, "http://localhost:4317", config.Endpoint)
	assert.Equal(t, TracingServiceName, config.ServiceName)
	assert.Equal(t, TracingServiceVersion, config.ServiceVersion)
	assert.Equal(t, 0.1, config.SamplingRatio)
	assert.True(t, config.Insecure)
	assert.NotNil(t, config.Headers)
	assert.Empty(t, config.Headers)
}

func TestInitTracingDisabled(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		Enabled: false,
	}

	tp, err := InitTracing(ctx, config)
	require.NoError(t, err)
	assert.Nil(t, tp, "TracerProvider should be nil when tracing is disabled")

	// Verify that a no-op tracer is set
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")
	assert.False(t, span.IsRecording(), "Span should not be recording when tracing is disabled")
	span.End()
}

func TestInitTracingNilConfig(t *testing.T) {
	ctx := context.Background()

	tp, err := InitTracing(ctx, nil)
	require.NoError(t, err)
	assert.Nil(t, tp, "TracerProvider should be nil with default disabled config")
}

func TestInitTracingEnabled(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		Enabled:       true,
		Endpoint:      "http://localhost:4317",
		ServiceName:   "test-service",
		ServiceVersion: "test-version",
		SamplingRatio: 1.0, // 100% sampling for testing
		Insecure:      true,
		Headers:       map[string]string{"test": "header"},
	}

	// Note: This test requires a running OTLP collector, so we expect an error
	// but we can still test the configuration parsing
	_, err := InitTracing(ctx, config)
	// We expect this to fail because there's no OTLP collector running
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create OTLP exporter")
	}
	// If no error, that means we successfully connected (unexpected but OK)
}

func TestTracer(t *testing.T) {
	tracer := Tracer()
	assert.NotNil(t, tracer)
	// Note: Cannot easily test instrumentation scope without casting to internal types
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()

	newCtx, span := StartSpan(ctx, "test-span")
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)

	// Verify span is in context
	spanFromCtx := trace.SpanFromContext(newCtx)
	assert.Equal(t, span, spanFromCtx)

	span.End()
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()

	// Test with no span in context
	span := SpanFromContext(ctx)
	assert.NotNil(t, span)
	assert.False(t, span.IsRecording())

	// Test with span in context
	_, testSpan := StartSpan(ctx, "test")
	spanCtx := trace.ContextWithSpan(ctx, testSpan)
	retrievedSpan := SpanFromContext(spanCtx)
	assert.Equal(t, testSpan, retrievedSpan)

	testSpan.End()
}

func TestStartOperation(t *testing.T) {
	ctx := context.Background()

	op := StartOperation(ctx, "test-operation",
		attribute.String("test.key", "test.value"),
		attribute.Int("test.number", 42),
	)

	assert.NotNil(t, op)
	assert.NotNil(t, op.Context())
	assert.NotEqual(t, ctx, op.Context(), "Operation context should be different from input context")

	op.End()
}

func TestOperationSetAttribute(t *testing.T) {
	ctx := context.Background()
	op := StartOperation(ctx, "test-operation")

	// Test different attribute types
	op.SetAttribute("string.value", "test")
	op.SetAttribute("int.value", 42)
	op.SetAttribute("int64.value", int64(123))
	op.SetAttribute("float64.value", 3.14)
	op.SetAttribute("bool.value", true)
	op.SetAttribute("other.value", struct{ Name string }{Name: "test"})

	// No assertions possible without instrumentation, but this tests the type handling
	op.End()
}

func TestOperationSetStatus(t *testing.T) {
	ctx := context.Background()
	op := StartOperation(ctx, "test-operation")

	op.SetStatus(codes.Ok, "Operation completed successfully")
	op.SetStatus(codes.Error, "Operation failed")

	op.End()
}

func TestOperationAddEvent(t *testing.T) {
	ctx := context.Background()
	op := StartOperation(ctx, "test-operation")

	op.AddEvent("test.event",
		attribute.String("event.data", "test"),
		attribute.Int("event.sequence", 1),
	)

	op.End()
}

func TestOperationRecordError(t *testing.T) {
	ctx := context.Background()
	op := StartOperation(ctx, "test-operation")

	// Test with nil error (should be safe)
	op.RecordError(nil)

	// Test with actual error
	testErr := assert.AnError
	op.RecordError(testErr,
		attribute.String("error.context", "test context"),
		attribute.String("error.component", "test component"),
	)

	op.End()
}

func TestOperationEndWithError(t *testing.T) {
	ctx := context.Background()
	op := StartOperation(ctx, "test-operation")

	// Test with nil error
	op2 := StartOperation(ctx, "test-operation-2")
	op2.EndWithError(nil)

	// Test with actual error
	testErr := assert.AnError
	op.EndWithError(testErr)
}

func TestShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should not panic with no tracer provider
	err := Shutdown(ctx)
	assert.NoError(t, err)
}

func TestNewTraceableHTTPClient(t *testing.T) {
	client := NewTraceableHTTPClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.tracer)
}

func TestAttributeKeys(t *testing.T) {
	// Test that all attribute keys are properly defined
	assert.Equal(t, "crossplane.resource.type", string(AttrResourceType))
	assert.Equal(t, "crossplane.resource.name", string(AttrResourceName))
	assert.Equal(t, "crossplane.operation", string(AttrOperation))
	assert.Equal(t, "crossplane.provider_config", string(AttrProviderConfig))
	assert.Equal(t, "mailgun.domain", string(AttrDomain))
	assert.Equal(t, "mailgun.credential.type", string(AttrCredentialType))
	assert.Equal(t, "mailgun.api.endpoint", string(AttrAPIEndpoint))
	assert.Equal(t, "http.method", string(AttrHTTPMethod))
	assert.Equal(t, "http.status_code", string(AttrHTTPStatusCode))
	assert.Equal(t, "retry.attempt", string(AttrRetryAttempt))
	assert.Equal(t, "circuit_breaker.state", string(AttrCircuitState))
}

func TestSpanNames(t *testing.T) {
	// Test that all span names are properly defined
	assert.Equal(t, "crossplane.resource.reconcile", SpanResourceReconcile)
	assert.Equal(t, "crossplane.resource.observe", SpanResourceObserve)
	assert.Equal(t, "crossplane.resource.create", SpanResourceCreate)
	assert.Equal(t, "crossplane.resource.update", SpanResourceUpdate)
	assert.Equal(t, "crossplane.resource.delete", SpanResourceDelete)
	assert.Equal(t, "mailgun.api.request", SpanMailgunAPI)
	assert.Equal(t, "kubernetes.secret.operation", SpanSecretOperation)
	assert.Equal(t, "resilience.retry", SpanRetryOperation)
	assert.Equal(t, "resilience.circuit_breaker", SpanCircuitBreaker)
}

func TestTracingConstants(t *testing.T) {
	assert.Equal(t, "provider-mailgun", TracingServiceName)
	assert.Equal(t, "v0.1.0", TracingServiceVersion)
	assert.Equal(t, "github.com/rossigee/provider-mailgun/internal/tracing", InstrumentationName)
}

func BenchmarkStartOperation(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op := StartOperation(ctx, "benchmark-operation")
		op.End()
	}
}

func BenchmarkSetAttribute(b *testing.B) {
	ctx := context.Background()
	op := StartOperation(ctx, "benchmark-operation")
	defer op.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op.SetAttribute("benchmark.iteration", i)
	}
}

func BenchmarkAddEvent(b *testing.B) {
	ctx := context.Background()
	op := StartOperation(ctx, "benchmark-operation")
	defer op.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op.AddEvent("benchmark.event", attribute.Int("iteration", i))
	}
}
