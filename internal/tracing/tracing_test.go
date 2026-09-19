package tracing

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		def   string
		setup func()
		tear  func()
		want  string
	}{
		{
			name:  "returns default when not set",
			key:   "TEST_GETENV_NOTSET",
			def:   "default-val",
			setup: func() {},
			tear:  func() {},
			want:  "default-val",
		},
		{
			name:  "returns value when set",
			key:   "TEST_GETENV_SET",
			def:   "default-val",
			setup: func() { _ = os.Setenv("TEST_GETENV_SET", "custom-val") },
			tear:  func() { _ = os.Unsetenv("TEST_GETENV_SET") },
			want:  "custom-val",
		},
		{
			name:  "empty string treated as unset returns default",
			key:   "TEST_GETENV_EMPTY",
			def:   "default-val",
			setup: func() { _ = os.Setenv("TEST_GETENV_EMPTY", "") },
			tear:  func() { _ = os.Unsetenv("TEST_GETENV_EMPTY") },
			want:  "default-val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tear()
			tt.setup()
			defer tt.tear()
			got := getEnv(tt.key, tt.def)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSpanAttrs(t *testing.T) {
	attrs := SpanAttrs("Domain", "my-domain", "observe")

	assert.Len(t, attrs, 3)
	assert.Contains(t, attrs, attribute.String(resourceTypeAttr, "Domain"))
	assert.Contains(t, attrs, attribute.String(resourceNameAttr, "my-domain"))
	assert.Contains(t, attrs, attribute.String(operationAttr, "observe"))
}

func TestStartSpan_NilTracer(t *testing.T) {
	oldTracer := tracer
	tracer = nil
	defer func() { tracer = oldTracer }()

	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")

	assert.Equal(t, ctx, newCtx)
	assert.Nil(t, span)
}

func TestStartSpanWithAttrs_NilContext(t *testing.T) {
	ctx, span := StartSpanWithAttrs(context.Background(), "test-op", "Domain", "my-domain", "create")

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestStartOperation_NilTracer(t *testing.T) {
	oldTracer := tracer
	tracer = nil
	defer func() { tracer = oldTracer }()

	op := StartOperation(context.Background(), "test-op")
	assert.NotNil(t, op)
	assert.Nil(t, op.span)
}

func TestOperation_End_NilSpan(t *testing.T) {
	op := &Operation{ctx: context.Background(), span: nil}
	op.End()
}

func TestOperation_SetAttribute(t *testing.T) {
	t.Run("string attribute", func(t *testing.T) {
		op := &Operation{}
		op.SetAttribute("key", "value")
	})

	t.Run("int attribute", func(t *testing.T) {
		op := &Operation{}
		op.SetAttribute("key", 42)
	})

	t.Run("int64 attribute", func(t *testing.T) {
		op := &Operation{}
		op.SetAttribute("key", int64(42))
	})

	t.Run("bool attribute", func(t *testing.T) {
		op := &Operation{}
		op.SetAttribute("key", true)
	})

	t.Run("float64 attribute", func(t *testing.T) {
		op := &Operation{}
		op.SetAttribute("key", 3.14)
	})
}

func TestOperation_RecordError_NilSpan(t *testing.T) {
	op := &Operation{ctx: context.Background(), span: nil}
	op.RecordError(nil)

	op.span = nil
	op.RecordError(context.DeadlineExceeded)
}

func TestInit_TracingDisabled(t *testing.T) {
	_ = os.Setenv("OTEL_TRACING_ENABLED", "false")
	defer func() { _ = os.Unsetenv("OTEL_TRACING_ENABLED") }()

	cleanup := Init("test-service")
	assert.NotNil(t, cleanup)
	cleanup(context.Background())
}

func TestInit_InvalidSamplingRatio(t *testing.T) {
	_ = os.Setenv("OTEL_TRACING_ENABLED", "true")
	_ = os.Setenv("OTEL_SAMPLING_RATIO", "not-a-number")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "invalid-endpoint")
	defer func() {
		_ = os.Unsetenv("OTEL_TRACING_ENABLED")
		_ = os.Unsetenv("OTEL_SAMPLING_RATIO")
		_ = os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}()

	cleanup := Init("test-service")
	assert.NotNil(t, cleanup)
	cleanup(context.Background())
}
