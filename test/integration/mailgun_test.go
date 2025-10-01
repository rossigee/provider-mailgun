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

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rossigee/provider-mailgun/internal/clients"
	"github.com/rossigee/provider-mailgun/internal/resilience"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
)

const (
	// Environment variables for integration testing
	EnvMailgunAPIKey  = "MAILGUN_API_KEY"
	EnvMailgunDomain  = "MAILGUN_TEST_DOMAIN"
	EnvMailgunBaseURL = "MAILGUN_BASE_URL"
	EnvSkipIntegration = "SKIP_INTEGRATION_TESTS"
)

// testError is a simple error type for testing
type testError struct {
	msg        string
	statusCode int
}

func (e *testError) Error() string {
	return e.msg
}

// newTestError creates a test error with message and status code
func newTestError(message string, statusCode int) error {
	return &testError{
		msg:        fmt.Sprintf("API request failed with status %d: %s", statusCode, message),
		statusCode: statusCode,
	}
}

// TestMailgunIntegration runs integration tests against real Mailgun API
// Set MAILGUN_API_KEY and MAILGUN_TEST_DOMAIN environment variables to run
func TestMailgunIntegration(t *testing.T) {
	if os.Getenv(EnvSkipIntegration) != "" {
		t.Skip("Integration tests skipped")
	}

	apiKey := os.Getenv(EnvMailgunAPIKey)
	testDomain := os.Getenv(EnvMailgunDomain)

	if apiKey == "" || testDomain == "" {
		t.Skip("Integration tests require MAILGUN_API_KEY and MAILGUN_TEST_DOMAIN environment variables")
	}

	baseURL := os.Getenv(EnvMailgunBaseURL)
	if baseURL == "" {
		baseURL = "https://api.mailgun.net/v3"
	}

	// Create client configuration
	config := &clients.Config{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: nil, // Use default HTTP client
	}

	// Create resilient client
	baseClient := clients.NewClient(config)
	client := resilience.NewResilientClient(baseClient, resilience.APIRetryConfig())

	t.Run("DomainOperations", func(t *testing.T) {
		testDomainOperations(t, client, testDomain)
	})

	t.Run("SMTPCredentialOperations", func(t *testing.T) {
		testSMTPCredentialOperations(t, client, testDomain)
	})

	t.Run("ResilienceFeatures", func(t *testing.T) {
		testResilienceFeatures(t, client, testDomain)
	})
}

func testDomainOperations(t *testing.T, client *resilience.ResilientClient, testDomain string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("GetDomain", func(t *testing.T) {
		domain, err := client.GetDomain(ctx, testDomain)

		if err != nil {
			// Domain might not exist - that's okay for this test
			if clients.IsNotFound(err) {
				t.Logf("Domain %s not found (this is expected if domain is not configured)", testDomain)
				return
			}
			t.Fatalf("Unexpected error getting domain: %v", err)
		}

		assert.NotNil(t, domain)
		// Note: Domain name is not returned in v1beta1 API observation - it's implicit from the request
		t.Logf("Domain found, State: %s", domain.State)
	})
}

func testSMTPCredentialOperations(t *testing.T, client *resilience.ResilientClient, testDomain string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	testLogin := "integration-test@" + testDomain

	// Cleanup function
	cleanup := func() {
		_ = client.DeleteSMTPCredential(ctx, testDomain, testLogin)
	}
	defer cleanup()

	t.Run("CreateSMTPCredential", func(t *testing.T) {
		spec := &smtpcredentialtypes.SMTPCredentialParameters{
			Login:    testLogin,
			Password: nil, // Let Mailgun generate password
		}

		credential, err := client.CreateSMTPCredential(ctx, testDomain, spec)

		if err != nil {
			if clients.IsNotFound(err) {
				t.Skipf("Domain %s not configured in Mailgun account", testDomain)
			}
			require.NoError(t, err, "Failed to create SMTP credential")
		}

		assert.NotNil(t, credential)
		assert.Equal(t, testLogin, credential.Login)
		// Note: Password is not returned in v1beta1 API for security reasons
		assert.NotEmpty(t, credential.CreatedAt, "CreatedAt should be set")

		t.Logf("Created SMTP credential: %s", credential.Login)
	})

	t.Run("GetSMTPCredential", func(t *testing.T) {
		// Note: SMTP credentials are write-only, so this should return 404
		_, err := client.GetSMTPCredential(ctx, testDomain, testLogin)

		// We expect this to fail because SMTP credentials are write-only
		assert.Error(t, err, "SMTP credentials should be write-only")
		assert.True(t, clients.IsNotFound(err), "Should return not found for SMTP credentials")

		t.Logf("Confirmed SMTP credentials are write-only (expected 404)")
	})

	t.Run("DeleteSMTPCredential", func(t *testing.T) {
		err := client.DeleteSMTPCredential(ctx, testDomain, testLogin)

		if err != nil && !clients.IsNotFound(err) {
			t.Errorf("Failed to delete SMTP credential: %v", err)
		}

		t.Logf("Deleted SMTP credential (or it didn't exist)")
	})
}

func testResilienceFeatures(t *testing.T, client *resilience.ResilientClient, testDomain string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("RetryOnNotFound", func(t *testing.T) {
		// Try to get a non-existent domain - should not retry 404 errors
		start := time.Now()
		_, err := client.GetDomain(ctx, "non-existent-domain-12345.test")
		duration := time.Since(start)

		assert.Error(t, err)
		assert.True(t, clients.IsNotFound(err))
		// Should not have taken long since 404s are not retried
		assert.Less(t, duration, 5*time.Second, "Should not retry on 404 errors")

		t.Logf("Non-retryable error handled correctly in %v", duration)
	})

	t.Run("AuthenticationError", func(t *testing.T) {
		// Create a client with invalid API key
		invalidConfig := &clients.Config{
			APIKey:  "invalid-api-key",
			BaseURL: "https://api.mailgun.net/v3",
		}

		invalidClient := resilience.NewResilientClient(
			clients.NewClient(invalidConfig),
			resilience.APIRetryConfig(),
		)

		start := time.Now()
		_, err := invalidClient.GetDomain(ctx, testDomain)
		duration := time.Since(start)

		assert.Error(t, err)
		// Should not have taken long since auth errors are not retried
		assert.Less(t, duration, 5*time.Second, "Should not retry on auth errors")

		t.Logf("Authentication error handled correctly in %v", duration)
	})
}

// BenchmarkMailgunOperations runs performance benchmarks
func BenchmarkMailgunOperations(b *testing.B) {
	if os.Getenv(EnvSkipIntegration) != "" || os.Getenv(EnvMailgunAPIKey) == "" {
		b.Skip("Integration benchmarks require API key")
	}

	apiKey := os.Getenv(EnvMailgunAPIKey)
	testDomain := os.Getenv(EnvMailgunDomain)

	if apiKey == "" || testDomain == "" {
		b.Skip("Benchmarks require MAILGUN_API_KEY and MAILGUN_TEST_DOMAIN")
	}

	config := &clients.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.mailgun.net/v3",
	}

	baseClient := clients.NewClient(config)
	client := resilience.NewResilientClient(baseClient, resilience.APIRetryConfig())

	b.Run("GetDomain", func(b *testing.B) {
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = client.GetDomain(ctx, testDomain)
		}
	})

	b.Run("GetNonExistentDomain", func(b *testing.B) {
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = client.GetDomain(ctx, "non-existent-domain.test")
		}
	})
}

// TestClientConfiguration tests different client configurations
func TestClientConfiguration(t *testing.T) {
	t.Run("DefaultRetryConfig", func(t *testing.T) {
		config := resilience.DefaultRetryConfig()

		assert.Equal(t, 3, config.MaxAttempts)
		assert.Equal(t, 250*time.Millisecond, config.InitialBackoff)
		assert.Equal(t, 16*time.Second, config.MaxBackoff)
		assert.Equal(t, 0.1, config.BackoffJitter)
		assert.NotEmpty(t, config.RetryableErrors)
	})

	t.Run("APIRetryConfig", func(t *testing.T) {
		config := resilience.APIRetryConfig()

		assert.Equal(t, 5, config.MaxAttempts)
		assert.Equal(t, 500*time.Millisecond, config.InitialBackoff)
		assert.Equal(t, 30*time.Second, config.MaxBackoff)
		assert.Equal(t, 0.2, config.BackoffJitter)
		assert.Contains(t, config.RetryableErrors, "timeout")
		assert.Contains(t, config.RetryableErrors, "503 service unavailable")
	})

	t.Run("RetryableErrorDetection", func(t *testing.T) {
		config := resilience.APIRetryConfig()

		// Test retryable errors
		assert.True(t, config.IsRetryableError(newTestError("connection timeout", 0)))
		assert.True(t, config.IsRetryableError(newTestError("service unavailable", 503)))
		assert.True(t, config.IsRetryableError(newTestError("too many requests", 429)))

		// Test non-retryable errors
		assert.False(t, config.IsRetryableError(newTestError("not found", 404)))
		assert.False(t, config.IsRetryableError(newTestError("unauthorized", 401)))
		assert.False(t, config.IsRetryableError(newTestError("bad request", 400)))
	})

	t.Run("BackoffCalculation", func(t *testing.T) {
		config := &resilience.RetryConfig{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
			BackoffJitter:  0,
		}

		// Test exponential backoff
		assert.Equal(t, 100*time.Millisecond, config.CalculateBackoff(0))
		assert.Equal(t, 200*time.Millisecond, config.CalculateBackoff(1))
		assert.Equal(t, 400*time.Millisecond, config.CalculateBackoff(2))

		// Test max backoff limit
		backoff := config.CalculateBackoff(10) // Should be capped
		assert.LessOrEqual(t, backoff, 5*time.Second)
	})
}
