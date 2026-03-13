package services_test

import (
	conf_models "jellyfin-duplicate/configuration/models"
	"jellyfin-duplicate/configuration/services"
	"testing"

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				services.ConfigureLogrus(tt.config)
			}, "ConfigureLogrus should not panic with valid config")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				services.ConfigureGINMode()
			}, "ConfigureGINMode should not panic")
		})
	}
}
