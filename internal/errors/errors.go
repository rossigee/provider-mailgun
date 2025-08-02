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
	"strings"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Infrastructure errors
	ErrorCodeProviderConfig     ErrorCode = "ProviderConfigError"
	ErrorCodeAuthentication     ErrorCode = "AuthenticationError"
	ErrorCodeNetworkTimeout     ErrorCode = "NetworkTimeoutError"
	ErrorCodeRateLimited        ErrorCode = "RateLimitedError"
	ErrorCodeServiceUnavailable ErrorCode = "ServiceUnavailableError"

	// Resource errors
	ErrorCodeResourceNotFound   ErrorCode = "ResourceNotFoundError"
	ErrorCodeResourceConflict   ErrorCode = "ResourceConflictError"
	ErrorCodeInvalidSpec        ErrorCode = "InvalidSpecError"
	ErrorCodeValidationFailed   ErrorCode = "ValidationFailedError"

	// SMTP specific errors
	ErrorCodeDomainNotConfigured ErrorCode = "DomainNotConfiguredError"
	ErrorCodeCredentialRotation  ErrorCode = "CredentialRotationError"
	ErrorCodeSecretAccess        ErrorCode = "SecretAccessError"

	// Generic errors
	ErrorCodeInternal           ErrorCode = "InternalError"
	ErrorCodeUnknown           ErrorCode = "UnknownError"
)

// ProviderError represents a structured error with troubleshooting information
type ProviderError struct {
	Code            ErrorCode
	Message         string
	Cause           error
	TroubleshootURL string
	SuggestedAction string
	RetryAfter      string
	ResourceRef     string
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("[%s] %s", e.Code, e.Message))

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("Cause: %s", e.Cause.Error()))
	}

	if e.SuggestedAction != "" {
		parts = append(parts, fmt.Sprintf("Suggested Action: %s", e.SuggestedAction))
	}

	if e.TroubleshootURL != "" {
		parts = append(parts, fmt.Sprintf("Troubleshooting: %s", e.TroubleshootURL))
	}

	return strings.Join(parts, ". ")
}

// Unwrap returns the underlying cause
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// GetConditionReason returns an appropriate condition reason
func (e *ProviderError) GetConditionReason() string {
	switch e.Code {
	case ErrorCodeAuthentication, ErrorCodeProviderConfig:
		return "ConfigurationError"
	case ErrorCodeNetworkTimeout, ErrorCodeServiceUnavailable, ErrorCodeRateLimited:
		return "Unavailable"
	case ErrorCodeResourceNotFound:
		return "NotFound"
	case ErrorCodeValidationFailed:
		return "ValidationFailed"
	default:
		return "Error"
	}
}

// NewProviderError creates a new provider error
func NewProviderError(code ErrorCode, message string, cause error) *ProviderError {
	return &ProviderError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WithSuggestedAction adds a suggested action to the error
func (e *ProviderError) WithSuggestedAction(action string) *ProviderError {
	e.SuggestedAction = action
	return e
}

// WithTroubleshootURL adds a troubleshooting URL to the error
func (e *ProviderError) WithTroubleshootURL(url string) *ProviderError {
	e.TroubleshootURL = url
	return e
}

// WithRetryAfter adds retry timing information
func (e *ProviderError) WithRetryAfter(duration string) *ProviderError {
	e.RetryAfter = duration
	return e
}

// WithResourceRef adds resource reference information
func (e *ProviderError) WithResourceRef(ref string) *ProviderError {
	e.ResourceRef = ref
	return e
}

// Common error constructors with built-in troubleshooting

// NewAuthenticationError creates an authentication error with troubleshooting
func NewAuthenticationError(cause error) *ProviderError {
	return NewProviderError(
		ErrorCodeAuthentication,
		"Failed to authenticate with Mailgun API",
		cause,
	).WithSuggestedAction(
		"Verify your API key in the ProviderConfig secret. Ensure the key has sufficient permissions",
	).WithTroubleshootURL(
		"https://documentation.mailgun.com/en/latest/api-intro.html#authentication",
	)
}

// NewDomainNotConfiguredError creates a domain configuration error
func NewDomainNotConfiguredError(domain string) *ProviderError {
	return NewProviderError(
		ErrorCodeDomainNotConfigured,
		fmt.Sprintf("Domain '%s' is not configured in Mailgun", domain),
		nil,
	).WithSuggestedAction(
		"Add the domain to your Mailgun account or verify the domain name spelling",
	).WithTroubleshootURL(
		"https://documentation.mailgun.com/en/latest/user_manual.html#domains",
	)
}

// NewCredentialRotationError creates a credential rotation error
func NewCredentialRotationError(domain, login string, cause error) *ProviderError {
	return NewProviderError(
		ErrorCodeCredentialRotation,
		fmt.Sprintf("Failed to rotate SMTP credentials for %s@%s", login, domain),
		cause,
	).WithSuggestedAction(
		"Check if the domain exists and you have permission to manage SMTP credentials",
	).WithResourceRef(
		fmt.Sprintf("%s@%s", login, domain),
	)
}

// NewSecretAccessError creates a secret access error
func NewSecretAccessError(secretName, namespace string, cause error) *ProviderError {
	return NewProviderError(
		ErrorCodeSecretAccess,
		fmt.Sprintf("Failed to access secret %s/%s", namespace, secretName),
		cause,
	).WithSuggestedAction(
		"Verify the secret exists and the provider has RBAC permissions to access it",
	).WithResourceRef(
		fmt.Sprintf("%s/%s", namespace, secretName),
	)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(retryAfter string) *ProviderError {
	return NewProviderError(
		ErrorCodeRateLimited,
		"API rate limit exceeded",
		nil,
	).WithSuggestedAction(
		"Reduce the number of concurrent operations or upgrade your Mailgun plan",
	).WithRetryAfter(retryAfter).WithTroubleshootURL(
		"https://documentation.mailgun.com/en/latest/api-intro.html#rate-limiting",
	)
}

// NewNetworkTimeoutError creates a network timeout error
func NewNetworkTimeoutError(operation string, cause error) *ProviderError {
	return NewProviderError(
		ErrorCodeNetworkTimeout,
		fmt.Sprintf("Network timeout during %s operation", operation),
		cause,
	).WithSuggestedAction(
		"Check network connectivity and firewall rules. The operation will be retried automatically",
	).WithRetryAfter("exponential backoff")
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(cause error) *ProviderError {
	return NewProviderError(
		ErrorCodeServiceUnavailable,
		"Mailgun service is temporarily unavailable",
		cause,
	).WithSuggestedAction(
		"This is typically temporary. Check Mailgun status page for known issues",
	).WithTroubleshootURL(
		"https://status.mailgun.com",
	).WithRetryAfter("exponential backoff")
}

// NewValidationError creates a validation error
func NewValidationError(field, reason string) *ProviderError {
	return NewProviderError(
		ErrorCodeValidationFailed,
		fmt.Sprintf("Validation failed for field '%s': %s", field, reason),
		nil,
	).WithSuggestedAction(
		"Review the resource specification and fix the validation error",
	)
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(resourceType, identifier string) *ProviderError {
	return NewProviderError(
		ErrorCodeResourceNotFound,
		fmt.Sprintf("%s '%s' not found", resourceType, identifier),
		nil,
	).WithSuggestedAction(
		"Verify the resource exists in Mailgun or check the identifier spelling",
	).WithResourceRef(identifier)
}

// ErrorFromHTTPResponse creates an appropriate error from HTTP response
func ErrorFromHTTPResponse(resp *http.Response, operation string) *ProviderError {
	if resp == nil {
		return NewProviderError(ErrorCodeInternal, "nil HTTP response", nil)
	}

	switch resp.StatusCode {
	case 401:
		return NewAuthenticationError(fmt.Errorf("HTTP %d", resp.StatusCode))
	case 404:
		return NewProviderError(ErrorCodeResourceNotFound, "Resource not found", nil)
	case 409:
		return NewProviderError(ErrorCodeResourceConflict, "Resource conflict", nil)
	case 429:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter == "" {
			retryAfter = "60 seconds"
		}
		return NewRateLimitError(retryAfter)
	case 502, 503, 504:
		return NewServiceUnavailableError(fmt.Errorf("HTTP %d", resp.StatusCode))
	default:
		return NewProviderError(
			ErrorCodeUnknown,
			fmt.Sprintf("HTTP %d during %s", resp.StatusCode, operation),
			nil,
		)
	}
}

// WrapError wraps an existing error with provider error context
func WrapError(err error, code ErrorCode, message string) *ProviderError {
	if err == nil {
		return nil
	}

	// If it's already a ProviderError, don't double-wrap
	if pe, ok := err.(*ProviderError); ok {
		return pe
	}

	return NewProviderError(code, message, err)
}

// IsProviderError checks if an error is a ProviderError
func IsProviderError(err error) bool {
	_, ok := err.(*ProviderError)
	return ok
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if pe, ok := err.(*ProviderError); ok {
		return pe.Code
	}
	return ErrorCodeUnknown
}

// IsRetryable checks if an error indicates a retryable condition
func IsRetryable(err error) bool {
	if pe, ok := err.(*ProviderError); ok {
		switch pe.Code {
		case ErrorCodeNetworkTimeout, ErrorCodeRateLimited, ErrorCodeServiceUnavailable:
			return true
		}
	}
	return false
}

// GetRetryAfter extracts retry timing from an error
func GetRetryAfter(err error) string {
	if pe, ok := err.(*ProviderError); ok {
		return pe.RetryAfter
	}
	return ""
}
