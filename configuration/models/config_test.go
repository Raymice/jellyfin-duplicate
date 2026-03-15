package models_test

import (
	"jellyfin-duplicate/configuration/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJellyfinConfig tests the JellyfinConfig struct
func TestJellyfinConfig(t *testing.T) {
	tests := []struct {
		name   string
		config models.JellyfinConfig
	}{
		{
			name: "Valid Jellyfin config",
			config: models.JellyfinConfig{
				URL:    "http://localhost:8096",
				APIKey: "test-api-key-12345",
				UserID: "admin-user-id",
			},
		},
		{
			name: "Config with HTTPS",
			config: models.JellyfinConfig{
				URL:    "https://jellyfin.example.com",
				APIKey: "secure-api-key",
				UserID: "user-id-123",
			},
		},
		{
			name: "Config with empty values",
			config: models.JellyfinConfig{
				URL:    "",
				APIKey: "",
				UserID: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.config)
			// Verify struct fields are accessible
			assert.IsType(t, "", tt.config.URL)
			assert.IsType(t, "", tt.config.APIKey)
			assert.IsType(t, "", tt.config.UserID)
		})
	}
}

// TestLogrusConfig tests the LogrusConfig struct
func TestLogrusConfig(t *testing.T) {
	tests := []struct {
		name   string
		config models.LogrusConfig
	}{
		{
			name: "Info level text format",
			config: models.LogrusConfig{
				Level:         "info",
				Format:        "text",
				DisableColors: false,
				ReportCaller:  false,
			},
		},
		{
			name: "Debug level JSON format",
			config: models.LogrusConfig{
				Level:         "debug",
				Format:        "json",
				DisableColors: true,
				ReportCaller:  true,
			},
		},
		{
			name: "Warn level with colors disabled",
			config: models.LogrusConfig{
				Level:         "warn",
				Format:        "text",
				DisableColors: true,
				ReportCaller:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.config.Level)
			assert.IsType(t, "", tt.config.Format)
			assert.IsType(t, false, tt.config.DisableColors)
			assert.IsType(t, false, tt.config.ReportCaller)
		})
	}
}

// TestConfig tests the Config struct
func TestConfig(t *testing.T) {
	tests := []struct {
		name   string
		config models.Config
	}{
		{
			name: "Full config",
			config: models.Config{
				ServerPort: "8080",
				Logrus: models.LogrusConfig{
					Level:  "info",
					Format: "text",
				},
				Jellyfin: models.JellyfinConfig{
					URL:    "http://localhost:8096",
					APIKey: "key",
					UserID: "user",
				},
			},
		},
		{
			name: "Empty config",
			config: models.Config{
				ServerPort: "",
				Logrus:     models.LogrusConfig{},
				Jellyfin:   models.JellyfinConfig{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.config)
			assert.IsType(t, "", tt.config.ServerPort)
		})
	}
}
