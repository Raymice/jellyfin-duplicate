package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnvironmentConstants tests the Environment type and constants
func TestEnvironmentConstants(t *testing.T) {
	tests := []struct {
		name     string
		env      Environment
		expected string
	}{
		{
			name:     "Development environment",
			env:      Development,
			expected: "development",
		},
		{
			name:     "Production environment",
			env:      Production,
			expected: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, Environment(tt.expected), tt.env)
		})
	}
}

// TestEnvironmentVariableConstants tests the environment variable names
func TestEnvironmentVariableConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "Jellyfin URL env variable",
			constant: EnvJellyfinURL,
			expected: "JELLYFIN_URL",
		},
		{
			name:     "Jellyfin API Key env variable",
			constant: EnvJellyfinAPIKey,
			expected: "JELLYFIN_API_KEY",
		},
		{
			name:     "Jellyfin Admin User ID env variable",
			constant: EnvJellyfinAdminUserID,
			expected: "JELLYFIN_ADMIN_USER_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
			assert.NotEmpty(t, tt.constant)
		})
	}
}

// TestEnvironmentTypeString tests environment type string representation
func TestEnvironmentTypeString(t *testing.T) {
	tests := []struct {
		name string
		env  Environment
	}{
		{
			name: "Dev environment value",
			env:  Development,
		},
		{
			name: "Prod environment value",
			env:  Production,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify we can convert to string
			envStr := string(tt.env)
			assert.NotEmpty(t, envStr)
		})
	}
}
