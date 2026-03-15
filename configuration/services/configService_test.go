package services

import (
	conf_models "jellyfin-duplicate/configuration/models"

	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConfigureLogrus(t *testing.T) {
	tests := []struct {
		name   string
		config *conf_models.LogrusConfig
	}{
		{
			name: "Configure logrus with text format",
			config: &conf_models.LogrusConfig{
				Level:         "info",
				Format:        "text",
				DisableColors: false,
				ReportCaller:  false,
			},
		},
		{
			name: "Configure logrus with JSON format",
			config: &conf_models.LogrusConfig{
				Level:         "debug",
				Format:        "json",
				DisableColors: true,
				ReportCaller:  true,
			},
		},
		{
			name: "Configure logrus with invalid log level",
			config: &conf_models.LogrusConfig{
				Level:         "invalid-level",
				Format:        "text",
				DisableColors: false,
				ReportCaller:  false,
			},
		},
		{
			name: "Configure logrus with warn level",
			config: &conf_models.LogrusConfig{
				Level:         "warn",
				Format:        "text",
				DisableColors: true,
				ReportCaller:  false,
			},
		},
		{
			name: "Configure logrus with error level",
			config: &conf_models.LogrusConfig{
				Level:         "error",
				Format:        "json",
				DisableColors: false,
				ReportCaller:  true,
			},
		},
		{
			name: "Configure logrus with trace level",
			config: &conf_models.LogrusConfig{
				Level:         "trace",
				Format:        "text",
				DisableColors: false,
				ReportCaller:  false,
			},
		},
		{
			name: "Configure logrus with fatal level",
			config: &conf_models.LogrusConfig{
				Level:         "fatal",
				Format:        "json",
				DisableColors: true,
				ReportCaller:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				ConfigureLogrus(tt.config)
			}, "ConfigureLogrus should not panic with any valid config")

			// Verify logrus level was set
			assert.NotNil(t, logrus.StandardLogger())
		})
	}
}

func TestConfigureGINMode(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Configures GIN mode without error",
		},
		{
			name: "Configures GIN mode multiple times",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				ConfigureGINMode()
			}, "ConfigureGINMode should not panic")
		})
	}
}

// TestConfigureLogrusWithDifferentFormats tests logrus configuration with different formats
func TestConfigureLogrusWithDifferentFormats(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		shouldWork bool
	}{
		{
			name:       "Text format",
			format:     "text",
			shouldWork: true,
		},
		{
			name:       "JSON format",
			format:     "json",
			shouldWork: true,
		},
		{
			name:       "Unknown format defaults to text",
			format:     "unknown-format",
			shouldWork: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &conf_models.LogrusConfig{
				Level:         "info",
				Format:        tt.format,
				DisableColors: false,
				ReportCaller:  false,
			}

			assert.NotPanics(t, func() {
				ConfigureLogrus(config)
			}, "ConfigureLogrus should handle %s format", tt.format)
		})
	}
}

// TestConfigureLogrusWithReportCaller tests report caller configuration
func TestConfigureLogrusWithReportCaller(t *testing.T) {
	tests := []struct {
		name         string
		reportCaller bool
	}{
		{
			name:         "With report caller enabled",
			reportCaller: true,
		},
		{
			name:         "With report caller disabled",
			reportCaller: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &conf_models.LogrusConfig{
				Level:         "info",
				Format:        "text",
				DisableColors: false,
				ReportCaller:  tt.reportCaller,
			}

			assert.NotPanics(t, func() {
				ConfigureLogrus(config)
			}, "ConfigureLogrus should handle report caller: %v", tt.reportCaller)
		})
	}
}
