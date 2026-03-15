package constants_test

import (
	"jellyfin-duplicate/constants"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnvironmentConstants tests the Environment type and constants
func TestEnvironmentConstants(t *testing.T) {
	tests := []struct {
		name     string
		env      constants.Environment
		expected string
	}{
		{
			name:     "Development environment",
			env:      constants.Development,
			expected: "development",
		},
		{
			name:     "Production environment",
			env:      constants.Production,
			expected: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, constants.Environment(tt.expected), tt.env)
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
			constant: constants.EnvJellyfinURL,
			expected: "JELLYFIN_URL",
		},
		{
			name:     "Jellyfin API Key env variable",
			constant: constants.EnvJellyfinAPIKey,
			expected: "JELLYFIN_API_KEY",
		},
		{
			name:     "Jellyfin Admin User ID env variable",
			constant: constants.EnvJellyfinAdminUserID,
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
		env  constants.Environment
	}{
		{
			name: "Dev environment value",
			env:  constants.Development,
		},
		{
			name: "Prod environment value",
			env:  constants.Production,
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
