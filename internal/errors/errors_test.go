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

package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderError(t *testing.T) {
	t.Run("BasicError", func(t *testing.T) {
		err := NewProviderError(ErrorCodeAuthentication, "Failed to authenticate", fmt.Errorf("invalid key"))

		assert.Equal(t, ErrorCodeAuthentication, err.Code)
		assert.Equal(t, "Failed to authenticate", err.Message)
		assert.NotNil(t, err.Cause)
		assert.Equal(t, "invalid key", err.Cause.Error())
	})

	t.Run("ErrorString", func(t *testing.T) {
		err := NewProviderError(ErrorCodeAuthentication, "Failed to authenticate", fmt.Errorf("invalid key")).
			WithSuggestedAction("Check your API key").
			WithTroubleshootURL("https://docs.mailgun.com/auth")

		errorStr := err.Error()
		assert.Contains(t, errorStr, "[AuthenticationError] Failed to authenticate")
		assert.Contains(t, errorStr, "Cause: invalid key")
		assert.Contains(t, errorStr, "Suggested Action: Check your API key")
		assert.Contains(t, errorStr, "Troubleshooting: https://docs.mailgun.com/auth")
	})

	t.Run("Unwrap", func(t *testing.T) {
		cause := fmt.Errorf("underlying error")
		err := NewProviderError(ErrorCodeInternal, "Wrapper error", cause)

		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("GetConditionReason", func(t *testing.T) {
		tests := []struct {
			code     ErrorCode
			expected string
		}{
			{ErrorCodeAuthentication, "ConfigurationError"},
			{ErrorCodeProviderConfig, "ConfigurationError"},
			{ErrorCodeNetworkTimeout, "Unavailable"},
			{ErrorCodeServiceUnavailable, "Unavailable"},
			{ErrorCodeRateLimited, "Unavailable"},
			{ErrorCodeResourceNotFound, "NotFound"},
			{ErrorCodeValidationFailed, "ValidationFailed"},
			{ErrorCodeUnknown, "Error"},
		}

		for _, tt := range tests {
			err := NewProviderError(tt.code, "test message", nil)
			assert.Equal(t, tt.expected, err.GetConditionReason())
		}
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewAuthenticationError", func(t *testing.T) {
		err := NewAuthenticationError(fmt.Errorf("401 Unauthorized"))

		assert.Equal(t, ErrorCodeAuthentication, err.Code)
		assert.Contains(t, err.Message, "Failed to authenticate with Mailgun API")
		assert.Contains(t, err.SuggestedAction, "Verify your API key")
		assert.Contains(t, err.TroubleshootURL, "mailgun.com")
	})

	t.Run("NewDomainNotConfiguredError", func(t *testing.T) {
		err := NewDomainNotConfiguredError("example.com")

		assert.Equal(t, ErrorCodeDomainNotConfigured, err.Code)
		assert.Contains(t, err.Message, "example.com")
		assert.Contains(t, err.SuggestedAction, "Add the domain")
	})

	t.Run("NewCredentialRotationError", func(t *testing.T) {
		err := NewCredentialRotationError("example.com", "user", fmt.Errorf("rotation failed"))

		assert.Equal(t, ErrorCodeCredentialRotation, err.Code)
		assert.Contains(t, err.Message, "user@example.com")
		assert.Equal(t, "user@example.com", err.ResourceRef)
	})

	t.Run("NewSecretAccessError", func(t *testing.T) {
		err := NewSecretAccessError("my-secret", "default", fmt.Errorf("not found"))

		assert.Equal(t, ErrorCodeSecretAccess, err.Code)
		assert.Contains(t, err.Message, "default/my-secret")
		assert.Equal(t, "default/my-secret", err.ResourceRef)
	})

	t.Run("NewRateLimitError", func(t *testing.T) {
		err := NewRateLimitError("60 seconds")

		assert.Equal(t, ErrorCodeRateLimited, err.Code)
		assert.Equal(t, "60 seconds", err.RetryAfter)
		assert.Contains(t, err.SuggestedAction, "Reduce the number")
	})

	t.Run("NewNetworkTimeoutError", func(t *testing.T) {
		err := NewNetworkTimeoutError("create domain", fmt.Errorf("timeout"))

		assert.Equal(t, ErrorCodeNetworkTimeout, err.Code)
		assert.Contains(t, err.Message, "create domain")
		assert.Equal(t, "exponential backoff", err.RetryAfter)
	})

	t.Run("NewServiceUnavailableError", func(t *testing.T) {
		err := NewServiceUnavailableError(fmt.Errorf("503"))

		assert.Equal(t, ErrorCodeServiceUnavailable, err.Code)
		assert.Contains(t, err.TroubleshootURL, "status.mailgun.com")
		assert.Equal(t, "exponential backoff", err.RetryAfter)
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("email", "invalid format")

		assert.Equal(t, ErrorCodeValidationFailed, err.Code)
		assert.Contains(t, err.Message, "email")
		assert.Contains(t, err.Message, "invalid format")
	})

	t.Run("NewResourceNotFoundError", func(t *testing.T) {
		err := NewResourceNotFoundError("Domain", "example.com")

		assert.Equal(t, ErrorCodeResourceNotFound, err.Code)
		assert.Contains(t, err.Message, "Domain 'example.com' not found")
		assert.Equal(t, "example.com", err.ResourceRef)
	})
}

func TestErrorFromHTTPResponse(t *testing.T) {
	tests := []struct {
		statusCode   int
		expectedCode ErrorCode
		operation    string
	}{
		{401, ErrorCodeAuthentication, "get domain"},
		{404, ErrorCodeResourceNotFound, "get domain"},
		{409, ErrorCodeResourceConflict, "create domain"},
		{429, ErrorCodeRateLimited, "create credential"},
		{502, ErrorCodeServiceUnavailable, "update webhook"},
		{503, ErrorCodeServiceUnavailable, "delete route"},
		{504, ErrorCodeServiceUnavailable, "get template"},
		{500, ErrorCodeUnknown, "unknown operation"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("HTTP%d", tt.statusCode), func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
			}

			if tt.statusCode == 429 {
				resp.Header.Set("Retry-After", "120")
			}

			err := ErrorFromHTTPResponse(resp, tt.operation)
			assert.Equal(t, tt.expectedCode, err.Code)

			if tt.statusCode == 429 {
				assert.Equal(t, "120", err.RetryAfter)
			}
		})
	}

	t.Run("NilResponse", func(t *testing.T) {
		err := ErrorFromHTTPResponse(nil, "test")
		assert.Equal(t, ErrorCodeInternal, err.Code)
		assert.Contains(t, err.Message, "nil HTTP response")
	})
}

func TestErrorUtilities(t *testing.T) {
	t.Run("WrapError", func(t *testing.T) {
		// Test wrapping regular error
		originalErr := fmt.Errorf("original error")
		wrappedErr := WrapError(originalErr, ErrorCodeInternal, "wrapped message")

		assert.Equal(t, ErrorCodeInternal, wrappedErr.Code)
		assert.Equal(t, "wrapped message", wrappedErr.Message)
		assert.Equal(t, originalErr, wrappedErr.Cause)

		// Test wrapping ProviderError (should not double-wrap)
		providerErr := NewProviderError(ErrorCodeAuthentication, "auth error", nil)
		reWrapped := WrapError(providerErr, ErrorCodeInternal, "should not wrap")

		assert.Equal(t, providerErr, reWrapped)

		// Test wrapping nil error
		nilWrapped := WrapError(nil, ErrorCodeInternal, "nil error")
		assert.Nil(t, nilWrapped)
	})

	t.Run("IsProviderError", func(t *testing.T) {
		providerErr := NewProviderError(ErrorCodeAuthentication, "test", nil)
		regularErr := fmt.Errorf("regular error")

		assert.True(t, IsProviderError(providerErr))
		assert.False(t, IsProviderError(regularErr))
		assert.False(t, IsProviderError(nil))
	})

	t.Run("GetErrorCode", func(t *testing.T) {
		providerErr := NewProviderError(ErrorCodeAuthentication, "test", nil)
		regularErr := fmt.Errorf("regular error")

		assert.Equal(t, ErrorCodeAuthentication, GetErrorCode(providerErr))
		assert.Equal(t, ErrorCodeUnknown, GetErrorCode(regularErr))
		assert.Equal(t, ErrorCodeUnknown, GetErrorCode(nil))
	})

	t.Run("IsRetryable", func(t *testing.T) {
		tests := []struct {
			code     ErrorCode
			expected bool
		}{
			{ErrorCodeNetworkTimeout, true},
			{ErrorCodeRateLimited, true},
			{ErrorCodeServiceUnavailable, true},
			{ErrorCodeAuthentication, false},
			{ErrorCodeResourceNotFound, false},
			{ErrorCodeValidationFailed, false},
		}

		for _, tt := range tests {
			err := NewProviderError(tt.code, "test", nil)
			assert.Equal(t, tt.expected, IsRetryable(err))
		}

		// Test with regular error
		assert.False(t, IsRetryable(fmt.Errorf("regular error")))
	})

	t.Run("GetRetryAfter", func(t *testing.T) {
		errWithRetry := NewProviderError(ErrorCodeRateLimited, "test", nil).WithRetryAfter("60 seconds")
		errWithoutRetry := NewProviderError(ErrorCodeAuthentication, "test", nil)
		regularErr := fmt.Errorf("regular error")

		assert.Equal(t, "60 seconds", GetRetryAfter(errWithRetry))
		assert.Equal(t, "", GetRetryAfter(errWithoutRetry))
		assert.Equal(t, "", GetRetryAfter(regularErr))
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("WithMethods", func(t *testing.T) {
		err := NewProviderError(ErrorCodeAuthentication, "auth failed", nil).
			WithSuggestedAction("Check API key").
			WithTroubleshootURL("https://docs.example.com").
			WithRetryAfter("60 seconds").
			WithResourceRef("my-resource")

		assert.Equal(t, "Check API key", err.SuggestedAction)
		assert.Equal(t, "https://docs.example.com", err.TroubleshootURL)
		assert.Equal(t, "60 seconds", err.RetryAfter)
		assert.Equal(t, "my-resource", err.ResourceRef)
	})
}
