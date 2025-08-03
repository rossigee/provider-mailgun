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
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	// TracingServiceName is the service name for tracing
	TracingServiceName = "provider-mailgun"

	// TracingServiceVersion is the service version for tracing
	TracingServiceVersion = "v0.1.0"

	// InstrumentationName is the name of this instrumentation package
	InstrumentationName = "github.com/rossigee/provider-mailgun/internal/tracing"
)

// Config holds tracing configuration
type Config struct {
	// Enabled controls whether tracing is enabled
	Enabled bool

	// Endpoint is the OTLP collector endpoint
	Endpoint string

	// ServiceName is the service name for traces
	ServiceName string

	// ServiceVersion is the service version for traces
	ServiceVersion string

	// SamplingRatio is the sampling ratio (0.0 to 1.0)
	SamplingRatio float64

	// Insecure controls whether to use insecure connection
	Insecure bool

	// Headers contains additional headers for OTLP export
	Headers map[string]string
}

// DefaultConfig returns a default tracing configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       false, // Disabled by default
		Endpoint:      "http://localhost:4317",
		ServiceName:   TracingServiceName,
		ServiceVersion: TracingServiceVersion,
		SamplingRatio: 0.1, // 10% sampling by default
		Insecure:      true,
		Headers:       make(map[string]string),
	}
}

// TracerProvider holds the global tracer provider
var globalTracerProvider trace.TracerProvider

// Tracer returns a tracer for the provider
func Tracer() trace.Tracer {
	return otel.Tracer(InstrumentationName)
}

// InitTracing initializes OpenTelemetry tracing
func InitTracing(ctx context.Context, config *Config) (*sdktrace.TracerProvider, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// If tracing is disabled, return a no-op provider
	if !config.Enabled {
		noopProvider := noop.NewTracerProvider()
		otel.SetTracerProvider(noopProvider)
		globalTracerProvider = noopProvider
		return nil, nil
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String("kubernetes"),
			attribute.String("provider.type", "crossplane"),
			attribute.String("provider.name", "mailgun"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	exporter, err := createOTLPExporter(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SamplingRatio)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)
	globalTracerProvider = tp

	// Set global text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// createOTLPExporter creates an OTLP trace exporter
func createOTLPExporter(ctx context.Context, config *Config) (sdktrace.SpanExporter, error) {
	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithTimeout(10 * time.Second),
	}

	if config.Insecure {
		options = append(options, otlptracegrpc.WithInsecure())
	}

	if len(config.Headers) > 0 {
		options = append(options, otlptracegrpc.WithHeaders(config.Headers))
	}

	client := otlptracegrpc.NewClient(options...)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	return exporter, nil
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context) error {
	if tp, ok := globalTracerProvider.(*sdktrace.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

// Operation represents a traced operation
type Operation struct {
	span trace.Span
	ctx  context.Context
}

// StartOperation starts a new traced operation
func StartOperation(ctx context.Context, operationName string, attrs ...attribute.KeyValue) *Operation {
	newCtx, span := StartSpan(ctx, operationName, trace.WithAttributes(attrs...))
	return &Operation{
		span: span,
		ctx:  newCtx,
	}
}

// Context returns the context for this operation
func (op *Operation) Context() context.Context {
	return op.ctx
}

// SetAttribute adds an attribute to the span
func (op *Operation) SetAttribute(key string, value interface{}) {
	switch v := value.(type) {
	case string:
		op.span.SetAttributes(attribute.String(key, v))
	case int:
		op.span.SetAttributes(attribute.Int(key, v))
	case int64:
		op.span.SetAttributes(attribute.Int64(key, v))
	case float64:
		op.span.SetAttributes(attribute.Float64(key, v))
	case bool:
		op.span.SetAttributes(attribute.Bool(key, v))
	default:
		op.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// SetStatus sets the status of the operation
func (op *Operation) SetStatus(code codes.Code, description string) {
	op.span.SetStatus(code, description)
}

// AddEvent adds an event to the span
func (op *Operation) AddEvent(name string, attrs ...attribute.KeyValue) {
	op.span.AddEvent(name, trace.WithAttributes(attrs...))
}

// RecordError records an error in the span
func (op *Operation) RecordError(err error, attrs ...attribute.KeyValue) {
	if err != nil {
		op.span.RecordError(err, trace.WithAttributes(attrs...))
		op.span.SetStatus(codes.Error, err.Error())
	}
}

// End finishes the operation
func (op *Operation) End() {
	op.span.End()
}

// EndWithError finishes the operation and records an error if present
func (op *Operation) EndWithError(err error) {
	if err != nil {
		op.RecordError(err)
	}
	op.span.End()
}

// Common attribute keys
var (
	AttrResourceType   = attribute.Key("crossplane.resource.type")
	AttrResourceName   = attribute.Key("crossplane.resource.name")
	AttrOperation      = attribute.Key("crossplane.operation")
	AttrProviderConfig = attribute.Key("crossplane.provider_config")
	AttrDomain         = attribute.Key("mailgun.domain")
	AttrCredentialType = attribute.Key("mailgun.credential.type")
	AttrAPIEndpoint    = attribute.Key("mailgun.api.endpoint")
	AttrHTTPMethod     = attribute.Key("http.method")
	AttrHTTPStatusCode = attribute.Key("http.status_code")
	AttrRetryAttempt   = attribute.Key("retry.attempt")
	AttrCircuitState   = attribute.Key("circuit_breaker.state")
)

// TraceableHTTPClient wraps an HTTP client with tracing
type TraceableHTTPClient struct {
	tracer trace.Tracer
}

// NewTraceableHTTPClient creates a new traceable HTTP client
func NewTraceableHTTPClient() *TraceableHTTPClient {
	return &TraceableHTTPClient{
		tracer: Tracer(),
	}
}

// Common span names
const (
	SpanResourceReconcile = "crossplane.resource.reconcile"
	SpanResourceObserve   = "crossplane.resource.observe"
	SpanResourceCreate    = "crossplane.resource.create"
	SpanResourceUpdate    = "crossplane.resource.update"
	SpanResourceDelete    = "crossplane.resource.delete"
	SpanMailgunAPI        = "mailgun.api.request"
	SpanSecretOperation   = "kubernetes.secret.operation"
	SpanRetryOperation    = "resilience.retry"
	SpanCircuitBreaker    = "resilience.circuit_breaker"
)
