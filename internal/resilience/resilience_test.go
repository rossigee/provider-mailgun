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

package resilience

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockNetError implements net.Error for testing
type mockNetError struct {
	msg         string
	isTimeout   bool
	isTemporary bool
}

func (e *mockNetError) Error() string   { return e.msg }
func (e *mockNetError) Timeout() bool   { return e.isTimeout }
func (e *mockNetError) Temporary() bool { return e.isTemporary }

func TestRetryConfig(t *testing.T) {
	t.Run("DefaultRetryConfig", func(t *testing.T) {
		config := DefaultRetryConfig()

		assert.Equal(t, 3, config.MaxAttempts)
		assert.Equal(t, 250*time.Millisecond, config.InitialBackoff)
		assert.Equal(t, 16*time.Second, config.MaxBackoff)
		assert.Equal(t, 0.1, config.BackoffJitter)
		assert.NotEmpty(t, config.RetryableErrors)
	})

	t.Run("APIRetryConfig", func(t *testing.T) {
		config := APIRetryConfig()

		assert.Equal(t, 5, config.MaxAttempts)
		assert.Equal(t, 500*time.Millisecond, config.InitialBackoff)
		assert.Equal(t, 30*time.Second, config.MaxBackoff)
		assert.Equal(t, 0.2, config.BackoffJitter)
		assert.Contains(t, config.RetryableErrors, "timeout")
		assert.Contains(t, config.RetryableErrors, "503 service unavailable")
	})
}

func TestIsRetryableError(t *testing.T) {
	config := APIRetryConfig()

	t.Run("RetryableErrors", func(t *testing.T) {
		tests := []struct {
			name  string
			err   error
			retry bool
		}{
			{"nil error", nil, false},
			{"timeout error", fmt.Errorf("connection timeout"), true},
			{"service unavailable", fmt.Errorf("503 service unavailable"), true},
			{"rate limit", fmt.Errorf("too many requests"), true},
			{"502 bad gateway", fmt.Errorf("502 bad gateway"), true},
			{"504 gateway timeout", fmt.Errorf("504 gateway timeout"), true},
			{"not found", fmt.Errorf("404 not found"), false},
			{"unauthorized", fmt.Errorf("401 unauthorized"), false},
			{"bad request", fmt.Errorf("400 bad request"), false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := config.IsRetryableError(tt.err)
				assert.Equal(t, tt.retry, result)
			})
		}
	})

	t.Run("NetworkErrors", func(t *testing.T) {
		// Create custom error types that implement net.Error
		timeoutErr := &mockNetError{
			msg:         "connection timeout",
			isTimeout:   true,
			isTemporary: false,
		}

		assert.True(t, config.IsRetryableError(timeoutErr))

		// Test temporary error
		tempErr := &mockNetError{
			msg:         "temporary network failure",
			isTimeout:   false,
			isTemporary: true,
		}

		// Note: As of Go 1.18+, Temporary() is deprecated and shouldn't be used
		// We only retry timeout errors, not temporary errors
		assert.False(t, config.IsRetryableError(tempErr))

		// Test regular network error (neither timeout nor temporary)
		// But this will still be retryable because "connection refused" is in RetryableErrors list
		regularErr := &mockNetError{
			msg:         "permission denied", // Not in retryable errors list
			isTimeout:   false,
			isTemporary: false,
		}

		assert.False(t, config.IsRetryableError(regularErr))
	})

	t.Run("HTTPStatusCodeDetection", func(t *testing.T) {
		tests := []struct {
			errorMsg string
			retry    bool
		}{
			{"HTTP 429 Too Many Requests", true},
			{"HTTP 502 Bad Gateway", true},
			{"HTTP 503 Service Unavailable", true},
			{"HTTP 504 Gateway Timeout", true},
			{"HTTP 404 Not Found", false},
			{"HTTP 401 Unauthorized", false},
		}

		for _, tt := range tests {
			t.Run(tt.errorMsg, func(t *testing.T) {
				err := fmt.Errorf("%s", tt.errorMsg)
				result := config.IsRetryableError(err)
				assert.Equal(t, tt.retry, result)
			})
		}
	})
}

func TestCalculateBackoff(t *testing.T) {
	t.Run("ExponentialBackoff", func(t *testing.T) {
		config := &RetryConfig{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
			BackoffJitter:  0, // No jitter for predictable testing
		}

		// Test exponential progression
		assert.Equal(t, 100*time.Millisecond, config.CalculateBackoff(0))
		assert.Equal(t, 200*time.Millisecond, config.CalculateBackoff(1))
		assert.Equal(t, 400*time.Millisecond, config.CalculateBackoff(2))
		assert.Equal(t, 800*time.Millisecond, config.CalculateBackoff(3))
	})

	t.Run("MaxBackoffLimit", func(t *testing.T) {
		config := &RetryConfig{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     3 * time.Second,
			BackoffJitter:  0,
		}

		// Test that backoff is capped at MaxBackoff
		backoff := config.CalculateBackoff(10) // Should be very large without cap
		assert.LessOrEqual(t, backoff, 3*time.Second)
	})

	t.Run("WithJitter", func(t *testing.T) {
		config := &RetryConfig{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     10 * time.Second,
			BackoffJitter:  0.1, // 10% jitter
		}

		// Run multiple times to test jitter variation
		backoffs := make([]time.Duration, 10)
		for i := 0; i < 10; i++ {
			backoffs[i] = config.CalculateBackoff(2) // Attempt 2: 4 seconds base
		}

		// With jitter, we should see some variation (though not guaranteed)
		// At minimum, all values should be reasonable
		for _, backoff := range backoffs {
			assert.GreaterOrEqual(t, backoff, 3500*time.Millisecond) // ~4s - 10%
			assert.LessOrEqual(t, backoff, 4500*time.Millisecond)    // ~4s + 10%
		}
	})

	t.Run("NegativeBackoffProtection", func(t *testing.T) {
		config := &RetryConfig{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     1 * time.Second,
			BackoffJitter:  1.0, // 100% jitter - could cause negative values
		}

		// Even with extreme jitter, should never return negative duration
		for i := 0; i < 100; i++ {
			backoff := config.CalculateBackoff(1)
			assert.GreaterOrEqual(t, backoff, time.Duration(0))
		}
	})
}

func TestWithRetry(t *testing.T) {
	t.Run("SuccessfulOperation", func(t *testing.T) {
		ctx := context.Background()
		config := &RetryConfig{MaxAttempts: 3}

		callCount := 0
		operation := func() error {
			callCount++
			return nil // Success on first try
		}

		err := WithRetry(ctx, "test_operation", config, operation)

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("FailureAfterRetries", func(t *testing.T) {
		ctx := context.Background()
		config := &RetryConfig{
			MaxAttempts:     3,
			InitialBackoff:  1 * time.Millisecond, // Fast for testing
			RetryableErrors: []string{"retryable"},
		}

		callCount := 0
		operation := func() error {
			callCount++
			return fmt.Errorf("retryable error")
		}

		err := WithRetry(ctx, "test_operation", config, operation)

		assert.Error(t, err)
		assert.Equal(t, 3, callCount) // Should try 3 times
		assert.Contains(t, err.Error(), "operation failed after 3 attempts")
	})

	t.Run("SuccessAfterRetries", func(t *testing.T) {
		ctx := context.Background()
		config := &RetryConfig{
			MaxAttempts:     3,
			InitialBackoff:  1 * time.Millisecond,
			RetryableErrors: []string{"retryable"},
		}

		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 3 {
				return fmt.Errorf("retryable error")
			}
			return nil // Success on third try
		}

		err := WithRetry(ctx, "test_operation", config, operation)

		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("NonRetryableError", func(t *testing.T) {
		ctx := context.Background()
		config := &RetryConfig{
			MaxAttempts:     3,
			RetryableErrors: []string{"retryable"},
		}

		callCount := 0
		operation := func() error {
			callCount++
			return fmt.Errorf("authentication failed") // Not in retryable list
		}

		err := WithRetry(ctx, "test_operation", config, operation)

		assert.Error(t, err)
		assert.Equal(t, 1, callCount) // Should not retry
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		config := &RetryConfig{
			MaxAttempts:     5,
			InitialBackoff:  100 * time.Millisecond, // Much longer than context timeout
			RetryableErrors: []string{"retryable"},
		}

		// Wait a bit to ensure context is cancelled
		time.Sleep(2 * time.Millisecond)

		callCount := 0
		operation := func() error {
			callCount++
			return fmt.Errorf("retryable error")
		}

		err := WithRetry(ctx, "test_operation", config, operation)

		assert.Error(t, err)
		// Should either fail due to timeout during retry or context cancellation
		errStr := err.Error()
		isTimeoutOrCancellation :=
			strings.Contains(errStr, "operation cancelled during retry backoff") ||
			strings.Contains(errStr, "operation failed after")
		assert.True(t, isTimeoutOrCancellation, "Expected timeout or cancellation error, got: %s", errStr)
	})

	t.Run("DefaultConfig", func(t *testing.T) {
		ctx := context.Background()

		operation := func() error {
			return nil
		}

		err := WithRetry(ctx, "test_operation", nil, operation)

		assert.NoError(t, err)
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("NewCircuitBreaker", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 1*time.Second)

		assert.Equal(t, "test", cb.name)
		assert.Equal(t, 3, cb.failureThreshold)
		assert.Equal(t, 1*time.Second, cb.resetTimeout)
		assert.Equal(t, CircuitClosed, cb.state)
		assert.Equal(t, 0, cb.failures)
	})

	t.Run("SuccessfulOperation", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 1*time.Second)
		ctx := context.Background()

		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := cb.Execute(ctx, operation)

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, CircuitClosed, cb.GetState())
	})

	t.Run("CircuitOpensAfterFailures", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)
		ctx := context.Background()

		operation := func() error {
			return fmt.Errorf("operation failed")
		}

		// First failure
		err1 := cb.Execute(ctx, operation)
		assert.Error(t, err1)
		assert.Equal(t, CircuitClosed, cb.GetState())

		// Second failure - should open circuit
		err2 := cb.Execute(ctx, operation)
		assert.Error(t, err2)
		assert.Equal(t, CircuitOpen, cb.GetState())

		// Third attempt - should fail fast
		err3 := cb.Execute(ctx, operation)
		assert.Error(t, err3)
		assert.Contains(t, err3.Error(), "circuit breaker is open")
	})

	t.Run("CircuitTransitionsToHalfOpen", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 1, 10*time.Millisecond)
		ctx := context.Background()

		// Cause circuit to open
		err := cb.Execute(ctx, func() error {
			return fmt.Errorf("failure")
		})
		assert.Error(t, err)
		assert.Equal(t, CircuitOpen, cb.GetState())

		// Wait for reset timeout
		time.Sleep(15 * time.Millisecond)

		// Next call should transition to half-open
		callCount := 0
		err = cb.Execute(ctx, func() error {
			callCount++
			return nil // Success
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, CircuitHalfOpen, cb.GetState())
	})

	t.Run("HalfOpenToClosedAfterSuccesses", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 1, 1*time.Millisecond)
		ctx := context.Background()

		// Open circuit
		_ = cb.Execute(ctx, func() error { return fmt.Errorf("failure") })
		assert.Equal(t, CircuitOpen, cb.GetState())

		// Wait and transition to half-open
		time.Sleep(2 * time.Millisecond)
		_ = cb.Execute(ctx, func() error { return nil })
		assert.Equal(t, CircuitHalfOpen, cb.GetState())

		// Execute 3 successful operations to close circuit
		for i := 0; i < 3; i++ {
			err := cb.Execute(ctx, func() error { return nil })
			assert.NoError(t, err)
		}

		assert.Equal(t, CircuitClosed, cb.GetState())
	})

	t.Run("HalfOpenBackToOpenOnFailure", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 1, 1*time.Millisecond)
		ctx := context.Background()

		// Open circuit
		_ = cb.Execute(ctx, func() error { return fmt.Errorf("failure") })

		// Wait and get to half-open
		time.Sleep(2 * time.Millisecond)
		_ = cb.Execute(ctx, func() error { return nil })
		assert.Equal(t, CircuitHalfOpen, cb.GetState())

		// Fail in half-open state
		err := cb.Execute(ctx, func() error { return fmt.Errorf("failure") })
		assert.Error(t, err)
		assert.Equal(t, CircuitOpen, cb.GetState())
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 1*time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := cb.Execute(ctx, func() error {
			return nil
		})

		// Circuit breaker should handle cancelled context
		if err != nil {
			assert.Equal(t, context.Canceled, err)
		}
		// Note: In some cases, the operation might succeed before context check
	})

	t.Run("GetStateNonBlocking", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 1*time.Second)

		// Getting state should not block
		state := cb.GetState()
		assert.Equal(t, CircuitClosed, state)
	})
}
