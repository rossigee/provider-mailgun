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
	"math"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// Retry metrics
	retryAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "provider_mailgun",
			Name:      "retry_attempts_total",
			Help:      "Total number of retry attempts",
		},
		[]string{"operation", "attempt", "result"},
	)

	retryBackoffDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "provider_mailgun",
			Name:      "retry_backoff_duration_seconds",
			Help:      "Duration of retry backoff in seconds",
			Buckets:   []float64{0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 16.0, 32.0},
		},
		[]string{"operation"},
	)

	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "provider_mailgun",
			Name:      "circuit_breaker_state",
			Help:      "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"operation"},
	)
)

func init() {
	metrics.Registry.MustRegister(retryAttempts, retryBackoffDuration, circuitBreakerState)
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffJitter   float64
	RetryableErrors []string
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 250 * time.Millisecond,
		MaxBackoff:     16 * time.Second,
		BackoffJitter:  0.1,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"connection reset",
			"temporary failure",
			"service unavailable",
			"too many requests",
			"rate limit",
		},
	}
}

// APIRetryConfig returns retry config optimized for API calls
func APIRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:    5,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     30 * time.Second,
		BackoffJitter:  0.2,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"connection reset",
			"temporary failure",
			"service unavailable",
			"too many requests",
			"rate limit",
			"502 bad gateway",
			"503 service unavailable",
			"504 gateway timeout",
		},
	}
}

// IsRetryableError checks if an error should trigger a retry
func (c *RetryConfig) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Check for network errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// Check for HTTP status codes in error message
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "502") ||
	   strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		return true
	}

	// Check configured retryable error strings
	for _, retryableErr := range c.RetryableErrors {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// CalculateBackoff calculates the backoff duration with jitter
func (c *RetryConfig) CalculateBackoff(attempt int) time.Duration {
	// Exponential backoff: initial * 2^attempt
	backoff := float64(c.InitialBackoff) * math.Pow(2, float64(attempt))

	// Apply max backoff limit
	if backoff > float64(c.MaxBackoff) {
		backoff = float64(c.MaxBackoff)
	}

	// Add jitter to prevent thundering herd
	if c.BackoffJitter > 0 {
		jitter := backoff * c.BackoffJitter * (rand.Float64()*2 - 1) // ±jitter%
		backoff += jitter
	}

	// Ensure positive duration
	if backoff < 0 {
		backoff = float64(c.InitialBackoff)
	}

	return time.Duration(backoff)
}

// WithRetry executes a function with retry logic
func WithRetry(ctx context.Context, operation string, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	logger := log.FromContext(ctx).WithValues("operation", operation)

	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()

		if err == nil {
			// Success!
			if attempt > 0 {
				logger.Info("operation succeeded after retries", "attempts", attempt+1)
				retryAttempts.WithLabelValues(operation, fmt.Sprintf("%d", attempt+1), "success").Inc()
			}
			return nil
		}

		lastErr = err
		retryAttempts.WithLabelValues(operation, fmt.Sprintf("%d", attempt+1), "failure").Inc()

		// Check if this is the last attempt
		if attempt == config.MaxAttempts-1 {
			logger.Error(err, "operation failed after all retry attempts", "maxAttempts", config.MaxAttempts)
			break
		}

		// Check if error is retryable
		if !config.IsRetryableError(err) {
			logger.Info("error is not retryable, aborting", "error", err.Error())
			break
		}

		// Calculate backoff and wait
		backoff := config.CalculateBackoff(attempt)
		logger.Info("retrying operation after backoff",
			"attempt", attempt+1,
			"backoff", backoff,
			"error", err.Error())

		retryBackoffDuration.WithLabelValues(operation).Observe(backoff.Seconds())

		// Wait for backoff duration or context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry backoff: %w", ctx.Err())
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name            string
	failureThreshold int
	resetTimeout    time.Duration
	state          CircuitBreakerState
	failures       int
	lastFailureTime time.Time
	successCount   int
	mutex          chan struct{} // Simple mutex using channel
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:            name,
		failureThreshold: failureThreshold,
		resetTimeout:    resetTimeout,
		state:          CircuitClosed,
		mutex:          make(chan struct{}, 1),
	}
	cb.mutex <- struct{}{} // Initialize mutex
	return cb
}

// Execute runs the operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Acquire mutex
	select {
	case <-cb.mutex:
		defer func() { cb.mutex <- struct{}{} }()
	case <-ctx.Done():
		return ctx.Err()
	}

	logger := log.FromContext(ctx).WithValues("circuitBreaker", cb.name)

	// Check current state
	now := time.Now()
	switch cb.state {
	case CircuitOpen:
		if now.Sub(cb.lastFailureTime) > cb.resetTimeout {
			// Transition to half-open
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			circuitBreakerState.WithLabelValues(cb.name).Set(2)
			logger.Info("circuit breaker transitioning to half-open")
		} else {
			circuitBreakerState.WithLabelValues(cb.name).Set(1)
			return fmt.Errorf("circuit breaker is open for %s", cb.name)
		}

	case CircuitHalfOpen:
		circuitBreakerState.WithLabelValues(cb.name).Set(2)

	case CircuitClosed:
		circuitBreakerState.WithLabelValues(cb.name).Set(0)
	}

	// Execute the function
	err := fn()

	if err != nil {
		cb.recordFailure(now, logger)
		return err
	}

	cb.recordSuccess(logger)
	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure(failureTime time.Time, logger logr.Logger) {
	cb.failures++
	cb.lastFailureTime = failureTime

	if cb.failures >= cb.failureThreshold {
		cb.state = CircuitOpen
		logger.Info("circuit breaker opened due to failures",
			"failures", cb.failures,
			"threshold", cb.failureThreshold)
		circuitBreakerState.WithLabelValues(cb.name).Set(1)
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess(logger logr.Logger) {
	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= 3 { // Require 3 successes to close
			cb.state = CircuitClosed
			cb.failures = 0
			logger.Info("circuit breaker closed after successful operations")
			circuitBreakerState.WithLabelValues(cb.name).Set(0)
		}
	case CircuitClosed:
		cb.failures = 0 // Reset failure count on success
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	select {
	case <-cb.mutex:
		defer func() { cb.mutex <- struct{}{} }()
		return cb.state
	default:
		return cb.state // Return current state without blocking
	}
}
