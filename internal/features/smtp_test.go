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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordPolicy_GenerateSecurePassword(t *testing.T) {
	tests := []struct {
		name     string
		policy   *PasswordPolicy
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid policy",
			policy: &PasswordPolicy{
				MinLength:        12,
				MaxLength:        20,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   true,
				ExcludeAmbiguous: true,
			},
			wantErr: false,
		},
		{
			name: "min length greater than max",
			policy: &PasswordPolicy{
				MinLength: 20,
				MaxLength: 12,
			},
			wantErr: true,
			errMsg:  "minimum length cannot be greater than maximum length",
		},
		{
			name: "min length with requirements",
			policy: &PasswordPolicy{
				MinLength:        8,
				MaxLength:        16,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   true,
			},
			wantErr: false,
		},
		{
			name: "no requirements",
			policy: &PasswordPolicy{
				MinLength: 8,
				MaxLength: 16,
			},
			wantErr: true,
			errMsg:  "no characters available for password generation",
		},
		{
			name: "custom character set",
			policy: &PasswordPolicy{
				MinLength:          8,
				MaxLength:          16,
				CustomCharacterSet: "abcdef123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := tt.policy.GenerateSecurePassword()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, password)
			assert.GreaterOrEqual(t, len(password), tt.policy.MinLength)
			assert.LessOrEqual(t, len(password), tt.policy.MaxLength)

			// Validate requirements
			if tt.policy.RequireUppercase {
				assert.Regexp(t, `[A-Z]`, password, "Password should contain uppercase letters")
			}
			if tt.policy.RequireLowercase {
				assert.Regexp(t, `[a-z]`, password, "Password should contain lowercase letters")
			}
			if tt.policy.RequireNumbers {
				assert.Regexp(t, `[0-9]`, password, "Password should contain numbers")
			}
			if tt.policy.RequireSymbols {
				assert.Regexp(t, `[!@#$%^&*()\-_=+\[\]{}|;:,.<>?]`, password, "Password should contain symbols")
			}
			if tt.policy.ExcludeAmbiguous {
				assert.NotRegexp(t, `[0O1l]`, password, "Password should not contain ambiguous characters")
			}
		})
	}
}

func TestDefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()

	assert.Equal(t, 16, policy.MinLength)
	assert.Equal(t, 64, policy.MaxLength)
	assert.True(t, policy.RequireUppercase)
	assert.True(t, policy.RequireLowercase)
	assert.True(t, policy.RequireNumbers)
	assert.True(t, policy.RequireSymbols)
	assert.True(t, policy.ExcludeAmbiguous)

	// Test that default policy can generate valid passwords
	password, err := policy.GenerateSecurePassword()
	require.NoError(t, err)
	assert.NotEmpty(t, password)
}

func TestIPAllowlistEntry(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	entry := IPAllowlistEntry{
		IP:          "192.168.1.0/24",
		Description: "Office network",
		CreatedAt:   now,
		ExpiresAt:   &expiresAt,
	}

	assert.Equal(t, "192.168.1.0/24", entry.IP)
	assert.Equal(t, "Office network", entry.Description)
	assert.Equal(t, now, entry.CreatedAt)
	assert.NotNil(t, entry.ExpiresAt)
	assert.Equal(t, expiresAt, *entry.ExpiresAt)
}

func TestRotationPolicy_DefaultValues(t *testing.T) {
	policy := DefaultRotationPolicy()

	assert.False(t, policy.Enabled)
	assert.Equal(t, 30*24*time.Hour, policy.RotationInterval)
	assert.Equal(t, 7*24*time.Hour, policy.OverlapPeriod)
	assert.Equal(t, 90*24*time.Hour, policy.MaxAge)
	assert.Equal(t, 7*24*time.Hour, policy.NotifyBeforeExpiry)
	assert.False(t, policy.AutomaticRotation)
}

func TestEnhancedSMTPCredential_Structure(t *testing.T) {
	now := time.Now()
	nextRotation := now.Add(24 * time.Hour)

	credential := &EnhancedSMTPCredential{
		IPAllowlist: []IPAllowlistEntry{
			{
				IP:          "192.168.1.0/24",
				Description: "Office network",
				CreatedAt:   now,
			},
		},
		RotationPolicy: DefaultRotationPolicy(),
		Tags: map[string]string{
			"environment": "production",
			"team":        "backend",
		},
		LastRotated:  &now,
		NextRotation: &nextRotation,
		UsageStats: &UsageStatistics{
			LastUsed:         &now,
			TotalConnections: 100,
			TotalEmailsSent:  500,
			LastConnectionIP: "192.168.1.1",
			FailedAttempts:   2,
		},
	}

	assert.Len(t, credential.IPAllowlist, 1)
	assert.Equal(t, "192.168.1.0/24", credential.IPAllowlist[0].IP)
	assert.NotNil(t, credential.RotationPolicy)
	assert.Equal(t, "production", credential.Tags["environment"])
	assert.Equal(t, "backend", credential.Tags["team"])
	assert.NotNil(t, credential.LastRotated)
	assert.NotNil(t, credential.NextRotation)
	assert.NotNil(t, credential.UsageStats)
	assert.Equal(t, int64(100), credential.UsageStats.TotalConnections)
	assert.Equal(t, int64(500), credential.UsageStats.TotalEmailsSent)
}

func TestUsageStatistics_Structure(t *testing.T) {
	now := time.Now()
	failedTime := now.Add(-1 * time.Hour)

	stats := &UsageStatistics{
		LastUsed:          &now,
		TotalConnections:  150,
		TotalEmailsSent:   750,
		LastConnectionIP:  "10.0.0.1",
		FailedAttempts:    3,
		LastFailedAttempt: &failedTime,
	}

	assert.Equal(t, now, *stats.LastUsed)
	assert.Equal(t, int64(150), stats.TotalConnections)
	assert.Equal(t, int64(750), stats.TotalEmailsSent)
	assert.Equal(t, "10.0.0.1", stats.LastConnectionIP)
	assert.Equal(t, int64(3), stats.FailedAttempts)
	assert.Equal(t, failedTime, *stats.LastFailedAttempt)
}

func TestNewSMTPCredentialManager(t *testing.T) {
	// Mock client - we can't import the actual client here without circular dependency
	// So we'll test with nil and just verify the structure
	manager := NewSMTPCredentialManager(nil)

	assert.NotNil(t, manager)
	assert.Nil(t, manager.client) // Expected since we passed nil
}

func TestPasswordPolicyCharacterSets(t *testing.T) {
	tests := []struct {
		name     string
		policy   *PasswordPolicy
		contains []string
		excludes []string
	}{
		{
			name: "require uppercase only",
			policy: &PasswordPolicy{
				MinLength:        8,
				MaxLength:        12,
				RequireUppercase: true,
			},
			contains: []string{"A", "B", "Z"},
		},
		{
			name: "require lowercase only",
			policy: &PasswordPolicy{
				MinLength:        8,
				MaxLength:        12,
				RequireLowercase: true,
			},
			contains: []string{"a", "b", "z"},
		},
		{
			name: "require numbers only",
			policy: &PasswordPolicy{
				MinLength:        8,
				MaxLength:        12,
				RequireNumbers:   true,
			},
			contains: []string{"2", "3", "9"},
		},
		{
			name: "require symbols only",
			policy: &PasswordPolicy{
				MinLength:        8,
				MaxLength:        12,
				RequireSymbols:   true,
			},
			contains: []string{"!", "@", "#"},
		},
		{
			name: "exclude ambiguous",
			policy: &PasswordPolicy{
				MinLength:        20,
				MaxLength:        30,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   true,
				ExcludeAmbiguous: true,
			},
			excludes: []string{"0", "O", "1", "l"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a password to test character sets
			password, err := tt.policy.GenerateSecurePassword()
			require.NoError(t, err)

			// For single requirement tests, check that ALL characters are from the required set
			if len(tt.contains) > 0 {
				// Check that password contains characters from the expected character set
				for _, char := range password {
					switch tt.name {
					case "require uppercase only":
						assert.True(t, char >= 'A' && char <= 'Z', "All characters should be uppercase")
					case "require lowercase only":
						assert.True(t, char >= 'a' && char <= 'z', "All characters should be lowercase")
					case "require numbers only":
						assert.True(t, char >= '0' && char <= '9', "All characters should be numbers")
					case "require symbols only":
						assert.Contains(t, "!@#$%^&*()-_=+[]{}|;:,.<>?", string(char), "All characters should be symbols")
					}
				}
			}

			for _, char := range tt.excludes {
				assert.NotContains(t, password, char, "Password should not contain excluded character: %s", char)
			}
		})
	}
}

func TestPasswordPolicyEdgeCases(t *testing.T) {
	t.Run("minimum length equals maximum length", func(t *testing.T) {
		policy := &PasswordPolicy{
			MinLength:        10,
			MaxLength:        10,
			RequireLowercase: true, // Need at least one requirement
		}

		password, err := policy.GenerateSecurePassword()
		require.NoError(t, err)
		assert.Len(t, password, 10)
	})

	t.Run("all requirements with minimal length", func(t *testing.T) {
		policy := &PasswordPolicy{
			MinLength:        4, // Minimum to satisfy all requirements
			MaxLength:        6,
			RequireUppercase: true,
			RequireLowercase: true,
			RequireNumbers:   true,
			RequireSymbols:   true,
		}

		password, err := policy.GenerateSecurePassword()
		require.NoError(t, err)

		assert.Regexp(t, `[A-Z]`, password)
		assert.Regexp(t, `[a-z]`, password)
		assert.Regexp(t, `[0-9]`, password)
		assert.Regexp(t, `[!@#$%^&*()\-_=+\[\]{}|;:,.<>?]`, password)
	})

	t.Run("custom character set only", func(t *testing.T) {
		policy := &PasswordPolicy{
			MinLength:          8,
			MaxLength:          12,
			CustomCharacterSet: "abc123",
		}

		password, err := policy.GenerateSecurePassword()
		require.NoError(t, err)

		// Should only contain characters from custom set
		for _, char := range password {
			assert.Contains(t, "abc123", string(char))
		}
	})
}

func BenchmarkPasswordGeneration(b *testing.B) {
	policy := DefaultPasswordPolicy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := policy.GenerateSecurePassword()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordGenerationComplex(b *testing.B) {
	policy := &PasswordPolicy{
		MinLength:        32,
		MaxLength:        64,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
		ExcludeAmbiguous: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := policy.GenerateSecurePassword()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestRotationPolicyTimingCalculations(t *testing.T) {
	now := time.Now()

	t.Run("should rotate based on interval", func(t *testing.T) {
		policy := &RotationPolicy{
			Enabled:          true,
			RotationInterval: 24 * time.Hour,
		}

		// Create a credential that was rotated 25 hours ago
		credential := &EnhancedSMTPCredential{
			LastRotated: func() *time.Time { t := now.Add(-25 * time.Hour); return &t }(),
		}

		// Should need rotation based on interval
		shouldRotate := credential.LastRotated != nil &&
			time.Since(*credential.LastRotated) > policy.RotationInterval
		assert.True(t, shouldRotate)
	})

	t.Run("should not rotate within interval", func(t *testing.T) {
		policy := &RotationPolicy{
			Enabled:          true,
			RotationInterval: 24 * time.Hour,
		}

		// Create a credential that was rotated 12 hours ago
		credential := &EnhancedSMTPCredential{
			LastRotated: func() *time.Time { t := now.Add(-12 * time.Hour); return &t }(),
		}

		// Should not need rotation yet
		shouldRotate := credential.LastRotated != nil &&
			time.Since(*credential.LastRotated) > policy.RotationInterval
		assert.False(t, shouldRotate)
	})
}
