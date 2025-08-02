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

package features

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/crossplane-contrib/provider-mailgun/internal/clients"
	"github.com/crossplane-contrib/provider-mailgun/internal/errors"
	"github.com/crossplane-contrib/provider-mailgun/internal/tracing"
)

// SMTPCredentialManager provides advanced SMTP credential management
type SMTPCredentialManager struct {
	client clients.Client
}

// NewSMTPCredentialManager creates a new SMTP credential manager
func NewSMTPCredentialManager(client clients.Client) *SMTPCredentialManager {
	return &SMTPCredentialManager{
		client: client,
	}
}

// PasswordPolicy defines password generation policies
type PasswordPolicy struct {
	// MinLength is the minimum password length
	MinLength int

	// MaxLength is the maximum password length
	MaxLength int

	// RequireUppercase requires uppercase letters
	RequireUppercase bool

	// RequireLowercase requires lowercase letters
	RequireLowercase bool

	// RequireNumbers requires numeric characters
	RequireNumbers bool

	// RequireSymbols requires special symbols
	RequireSymbols bool

	// ExcludeAmbiguous excludes ambiguous characters (0, O, l, 1, etc.)
	ExcludeAmbiguous bool

	// CustomCharacterSet allows specifying a custom character set
	CustomCharacterSet string
}

// DefaultPasswordPolicy returns a secure default password policy
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:        16,
		MaxLength:        64,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
		ExcludeAmbiguous: true,
	}
}

// IPAllowlistEntry represents an IP address or CIDR block in an allowlist
type IPAllowlistEntry struct {
	// IP is the IP address or CIDR block
	IP string

	// Description is a human-readable description
	Description string

	// CreatedAt is when this entry was added
	CreatedAt time.Time

	// ExpiresAt is when this entry expires (optional)
	ExpiresAt *time.Time
}

// RotationPolicy defines credential rotation policies
type RotationPolicy struct {
	// Enabled controls whether rotation is enabled
	Enabled bool

	// RotationInterval is how often to rotate credentials
	RotationInterval time.Duration

	// OverlapPeriod is how long old credentials remain valid
	OverlapPeriod time.Duration

	// MaxAge is the maximum age before forced rotation
	MaxAge time.Duration

	// NotifyBeforeExpiry is how long before expiry to send notifications
	NotifyBeforeExpiry time.Duration

	// AutomaticRotation enables automatic rotation
	AutomaticRotation bool
}

// DefaultRotationPolicy returns a default rotation policy
func DefaultRotationPolicy() *RotationPolicy {
	return &RotationPolicy{
		Enabled:            false, // Disabled by default
		RotationInterval:   30 * 24 * time.Hour, // 30 days
		OverlapPeriod:      7 * 24 * time.Hour,  // 7 days
		MaxAge:             90 * 24 * time.Hour, // 90 days
		NotifyBeforeExpiry: 7 * 24 * time.Hour,  // 7 days
		AutomaticRotation:  false,
	}
}

// EnhancedSMTPCredential represents an SMTP credential with advanced features
type EnhancedSMTPCredential struct {
	*clients.SMTPCredential

	// IPAllowlist contains allowed IP addresses/ranges
	IPAllowlist []IPAllowlistEntry

	// RotationPolicy defines rotation behavior
	RotationPolicy *RotationPolicy

	// Tags for organizational purposes
	Tags map[string]string

	// LastRotated is when the credential was last rotated
	LastRotated *time.Time

	// NextRotation is when the next rotation is scheduled
	NextRotation *time.Time

	// UsageStats contains usage statistics
	UsageStats *UsageStatistics
}

// UsageStatistics tracks credential usage
type UsageStatistics struct {
	// LastUsed is when the credential was last used
	LastUsed *time.Time

	// TotalConnections is the total number of connections
	TotalConnections int64

	// TotalEmailsSent is the total number of emails sent
	TotalEmailsSent int64

	// LastConnectionIP is the IP of the last connection
	LastConnectionIP string

	// FailedAttempts is the number of failed authentication attempts
	FailedAttempts int64

	// LastFailedAttempt is when the last failed attempt occurred
	LastFailedAttempt *time.Time
}

// GenerateSecurePassword generates a secure password according to the policy
func (p *PasswordPolicy) GenerateSecurePassword() (string, error) {
	if p.MinLength > p.MaxLength {
		return "", fmt.Errorf("minimum length cannot be greater than maximum length")
	}

	// Define character sets
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	symbols := "!@#$%^&*()-_=+[]{}|;:,.<>?"

	if p.ExcludeAmbiguous {
		lowercase = strings.ReplaceAll(lowercase, "l", "")
		uppercase = strings.ReplaceAll(uppercase, "O", "")
		numbers = strings.ReplaceAll(numbers, "0", "")
		numbers = strings.ReplaceAll(numbers, "1", "")
	}

	var charset string
	var required []string

	// Build character set and required characters
	if p.CustomCharacterSet != "" {
		charset = p.CustomCharacterSet
	} else {
		if p.RequireLowercase {
			charset += lowercase
			required = append(required, lowercase)
		}
		if p.RequireUppercase {
			charset += uppercase
			required = append(required, uppercase)
		}
		if p.RequireNumbers {
			charset += numbers
			required = append(required, numbers)
		}
		if p.RequireSymbols {
			charset += symbols
			required = append(required, symbols)
		}
	}

	if len(charset) == 0 {
		return "", fmt.Errorf("no characters available for password generation")
	}

	// Generate password length
	length := p.MinLength
	if p.MaxLength > p.MinLength {
		lengthRange := p.MaxLength - p.MinLength + 1
		lengthBig, err := rand.Int(rand.Reader, big.NewInt(int64(lengthRange)))
		if err != nil {
			return "", fmt.Errorf("failed to generate password length: %w", err)
		}
		length = p.MinLength + int(lengthBig.Int64())
	}

	password := make([]byte, length)

	// First, ensure required character types are included
	for i, reqCharset := range required {
		if i >= length {
			break
		}
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(reqCharset))))
		if err != nil {
			return "", fmt.Errorf("failed to select required character: %w", err)
		}
		password[i] = reqCharset[charIndex.Int64()]
	}

	// Fill the rest with random characters from the full charset
	for i := len(required); i < length; i++ {
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to select random character: %w", err)
		}
		password[i] = charset[charIndex.Int64()]
	}

	// Shuffle the password to avoid predictable patterns
	for i := length - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("failed to shuffle password: %w", err)
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
}

// ValidateIPAllowlist validates an IP allowlist entry
func ValidateIPAllowlist(entry IPAllowlistEntry) error {
	if entry.IP == "" {
		return errors.NewValidationError("ip", "IP address cannot be empty")
	}

	// Try to parse as IP address first
	if ip := net.ParseIP(entry.IP); ip != nil {
		return nil
	}

	// Try to parse as CIDR
	if _, _, err := net.ParseCIDR(entry.IP); err == nil {
		return nil
	}

	return errors.NewValidationError("ip", "invalid IP address or CIDR block")
}

// IsIPAllowed checks if an IP address is allowed by the allowlist
func IsIPAllowed(ip string, allowlist []IPAllowlistEntry) bool {
	if len(allowlist) == 0 {
		// If no allowlist is specified, allow all IPs
		return true
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	now := time.Now()

	for _, entry := range allowlist {
		// Check if entry has expired
		if entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
			continue
		}

		// Try exact IP match
		if entryIP := net.ParseIP(entry.IP); entryIP != nil {
			if clientIP.Equal(entryIP) {
				return true
			}
			continue
		}

		// Try CIDR match
		if _, network, err := net.ParseCIDR(entry.IP); err == nil {
			if network.Contains(clientIP) {
				return true
			}
		}
	}

	return false
}

// CreateCredentialWithFeatures creates an SMTP credential with advanced features
func (m *SMTPCredentialManager) CreateCredentialWithFeatures(
	ctx context.Context,
	domain string,
	spec *clients.SMTPCredentialSpec,
	policy *PasswordPolicy,
	allowlist []IPAllowlistEntry,
	rotationPolicy *RotationPolicy,
	tags map[string]string,
) (*EnhancedSMTPCredential, error) {

	op := tracing.StartOperation(ctx, "smtp.create_with_features",
		tracing.AttrDomain.String(domain),
		tracing.AttrCredentialType.String("enhanced"),
	)
	defer op.End()

	// Validate inputs
	if domain == "" {
		err := errors.NewValidationError("domain", "domain cannot be empty")
		op.RecordError(err)
		return nil, err
	}

	if spec == nil {
		err := errors.NewValidationError("spec", "credential spec cannot be nil")
		op.RecordError(err)
		return nil, err
	}

	// Validate IP allowlist
	for _, entry := range allowlist {
		if err := ValidateIPAllowlist(entry); err != nil {
			op.RecordError(err)
			return nil, err
		}
	}

	// Generate secure password if not provided
	if spec.Password == nil && policy != nil {
		password, err := policy.GenerateSecurePassword()
		if err != nil {
			op.RecordError(err)
			return nil, errors.NewProviderError(
				errors.ErrorCodeInternal,
				"Failed to generate secure password",
				err,
			).WithSuggestedAction("Check password policy configuration")
		}
		spec.Password = &password
		op.SetAttribute("password.generated", true)
	}

	// Create the credential
	credential, err := m.client.CreateSMTPCredential(ctx, domain, spec)
	if err != nil {
		op.RecordError(err)
		return nil, err
	}

	// Build enhanced credential
	enhanced := &EnhancedSMTPCredential{
		SMTPCredential: credential,
		IPAllowlist:    allowlist,
		RotationPolicy: rotationPolicy,
		Tags:           tags,
	}

	// Set rotation schedule if enabled
	if rotationPolicy != nil && rotationPolicy.Enabled {
		now := time.Now()
		enhanced.LastRotated = &now
		nextRotation := now.Add(rotationPolicy.RotationInterval)
		enhanced.NextRotation = &nextRotation
		op.SetAttribute("rotation.enabled", true)
		op.SetAttribute("rotation.next", nextRotation.Format(time.RFC3339))
	}

	op.SetAttribute("credential.login", credential.Login)
	op.SetAttribute("allowlist.entries", len(allowlist))

	return enhanced, nil
}

// RotateCredential rotates an SMTP credential
func (m *SMTPCredentialManager) RotateCredential(
	ctx context.Context,
	domain string,
	login string,
	policy *PasswordPolicy,
) (*EnhancedSMTPCredential, error) {

	op := tracing.StartOperation(ctx, "smtp.rotate_credential",
		tracing.AttrDomain.String(domain),
		tracing.AttrOperation.String("rotate"),
	)
	defer op.End()

	// Generate new password
	newPassword, err := policy.GenerateSecurePassword()
	if err != nil {
		op.RecordError(err)
		return nil, errors.NewProviderError(
			errors.ErrorCodeInternal,
			"Failed to generate new password during rotation",
			err,
		)
	}

	// Update the credential
	credential, err := m.client.UpdateSMTPCredential(ctx, domain, login, newPassword)
	if err != nil {
		op.RecordError(err)
		return nil, err
	}

	now := time.Now()
	enhanced := &EnhancedSMTPCredential{
		SMTPCredential: credential,
		LastRotated:    &now,
	}

	op.SetAttribute("credential.login", login)
	op.SetAttribute("rotation.timestamp", now.Format(time.RFC3339))

	return enhanced, nil
}

// ValidateCredentialAccess validates if access should be allowed based on IP and other factors
func (m *SMTPCredentialManager) ValidateCredentialAccess(
	ctx context.Context,
	credential *EnhancedSMTPCredential,
	clientIP string,
) error {

	op := tracing.StartOperation(ctx, "smtp.validate_access",
		tracing.AttrCredentialType.String("enhanced"),
		attribute.String("client.ip", clientIP),
	)
	defer op.End()

	// Check IP allowlist
	if !IsIPAllowed(clientIP, credential.IPAllowlist) {
		err := errors.NewProviderError(
			errors.ErrorCodeSecretAccess,
			fmt.Sprintf("IP address %s is not in the allowlist", clientIP),
			nil,
		).WithSuggestedAction("Add the IP address to the credential allowlist")

		op.RecordError(err)
		op.SetAttribute("access.denied", true)
		op.SetAttribute("access.reason", "ip_not_allowed")
		return err
	}

	// Check if credential needs rotation
	if credential.RotationPolicy != nil && credential.RotationPolicy.Enabled {
		if credential.NextRotation != nil && time.Now().After(*credential.NextRotation) {
			err := errors.NewProviderError(
				errors.ErrorCodeCredentialRotation,
				"Credential has expired and needs rotation",
				nil,
			).WithSuggestedAction("Rotate the credential to continue using it")

			op.RecordError(err)
			op.SetAttribute("access.denied", true)
			op.SetAttribute("access.reason", "credential_expired")
			return err
		}
	}

	op.SetAttribute("access.allowed", true)
	return nil
}

// LoginValidator validates SMTP login format
type LoginValidator struct {
	allowedDomains   []string
	forbiddenPatterns []*regexp.Regexp
}

// NewLoginValidator creates a new login validator
func NewLoginValidator() *LoginValidator {
	return &LoginValidator{
		allowedDomains:    []string{},
		forbiddenPatterns: []*regexp.Regexp{},
	}
}

// WithAllowedDomains sets allowed domains for logins
func (v *LoginValidator) WithAllowedDomains(domains ...string) *LoginValidator {
	v.allowedDomains = domains
	return v
}

// WithForbiddenPatterns adds forbidden patterns (as regex)
func (v *LoginValidator) WithForbiddenPatterns(patterns ...string) *LoginValidator {
	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			v.forbiddenPatterns = append(v.forbiddenPatterns, re)
		}
	}
	return v
}

// ValidateLogin validates an SMTP login
func (v *LoginValidator) ValidateLogin(login string) error {
	if login == "" {
		return errors.NewValidationError("login", "login cannot be empty")
	}

	// Check email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(login) {
		return errors.NewValidationError("login", "login must be a valid email address")
	}

	// Check allowed domains
	if len(v.allowedDomains) > 0 {
		parts := strings.Split(login, "@")
		if len(parts) != 2 {
			return errors.NewValidationError("login", "invalid email format")
		}

		domain := parts[1]
		allowed := false
		for _, allowedDomain := range v.allowedDomains {
			if domain == allowedDomain {
				allowed = true
				break
			}
		}

		if !allowed {
			return errors.NewValidationError("login",
				fmt.Sprintf("domain %s is not in the allowed domains list", domain))
		}
	}

	// Check forbidden patterns
	for _, pattern := range v.forbiddenPatterns {
		if pattern.MatchString(login) {
			return errors.NewValidationError("login",
				"login matches a forbidden pattern")
		}
	}

	return nil
}

// CredentialMetrics tracks credential usage and health
type CredentialMetrics struct {
	TotalCredentials       int64
	ActiveCredentials      int64
	ExpiredCredentials     int64
	RotationsLast30Days    int64
	FailedAuthsLast24Hours int64
	AverageCredentialAge   time.Duration
}

// GetCredentialMetrics returns metrics about credential usage
func (m *SMTPCredentialManager) GetCredentialMetrics(ctx context.Context) (*CredentialMetrics, error) {
	op := tracing.StartOperation(ctx, "smtp.get_metrics")
	defer op.End()

	// In a real implementation, this would query actual credential data
	// For now, return example metrics
	metrics := &CredentialMetrics{
		TotalCredentials:       100,
		ActiveCredentials:      85,
		ExpiredCredentials:     15,
		RotationsLast30Days:    12,
		FailedAuthsLast24Hours: 3,
		AverageCredentialAge:   45 * 24 * time.Hour,
	}

	op.SetAttribute("metrics.total", metrics.TotalCredentials)
	op.SetAttribute("metrics.active", metrics.ActiveCredentials)
	op.SetAttribute("metrics.expired", metrics.ExpiredCredentials)

	return metrics, nil
}
