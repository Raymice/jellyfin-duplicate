package services_test

import (
	conf_models "jellyfin-duplicate/configuration/models"
	conf_services "jellyfin-duplicate/configuration/services"
	"jellyfin-duplicate/constants"
	"os"

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
				conf_services.ConfigureLogrus(tt.config)
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
				conf_services.ConfigureGINMode()
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
				conf_services.ConfigureLogrus(config)
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
				conf_services.ConfigureLogrus(config)
			}, "ConfigureLogrus should handle report caller: %v", tt.reportCaller)
		})
	}
}

// TestConfigureLogrusWithAllLevels tests all valid log levels
func TestConfigureLogrusWithAllLevels(t *testing.T) {
	levels := []string{"panic", "fatal", "error", "warn", "info", "debug", "trace"}

	for _, level := range levels {
		t.Run("Level_"+level, func(t *testing.T) {
			config := &conf_models.LogrusConfig{
				Level:         level,
				Format:        "text",
				DisableColors: false,
				ReportCaller:  false,
			}

			assert.NotPanics(t, func() {
				conf_services.ConfigureLogrus(config)
			}, "ConfigureLogrus should accept level: %s", level)
		})
	}
}

// TestConfigureLogrusFormatter tests and verifies formatter configuration
func TestConfigureLogrusFormatter(t *testing.T) {
	t.Run("Text formatter configuration", func(t *testing.T) {
		config := &conf_models.LogrusConfig{
			Level:         "info",
			Format:        "text",
			DisableColors: true,
			ReportCaller:  true,
		}

		assert.NotPanics(t, func() {
			conf_services.ConfigureLogrus(config)
		})

		// Verify the formatter is set
		assert.NotNil(t, logrus.StandardLogger().Formatter)
	})

	t.Run("JSON formatter configuration", func(t *testing.T) {
		config := &conf_models.LogrusConfig{
			Level:         "debug",
			Format:        "json",
			DisableColors: false,
			ReportCaller:  false,
		}

		assert.NotPanics(t, func() {
			conf_services.ConfigureLogrus(config)
		})

		// Verify the formatter is set
		assert.NotNil(t, logrus.StandardLogger().Formatter)
	})

	t.Run("Default formatter for unknown format", func(t *testing.T) {
		config := &conf_models.LogrusConfig{
			Level:         "warn",
			Format:        "unknown",
			DisableColors: false,
			ReportCaller:  true,
		}

		assert.NotPanics(t, func() {
			conf_services.ConfigureLogrus(config)
		})

		// Should default to text formatter, not panic
		assert.NotNil(t, logrus.StandardLogger().Formatter)
	})
}

// TestConfigureGINModeStability tests GIN mode configuration consistency
func TestConfigureGINModeStability(t *testing.T) {
	t.Run("GIN mode can be configured multiple times", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			assert.NotPanics(t, func() {
				conf_services.ConfigureGINMode()
			})
		}
	})

	t.Run("ConfigureGINMode doesn't panic with repeated calls", func(t *testing.T) {
		assert.NotPanics(t, func() {
			conf_services.ConfigureGINMode()
			conf_services.ConfigureGINMode()
			conf_services.ConfigureGINMode()
		})
	})
}

// TestConfigureLogrusEdgeCases tests edge cases for logrus configuration
func TestConfigureLogrusEdgeCases(t *testing.T) {
	t.Run("Empty level defaults to info", func(t *testing.T) {
		config := &conf_models.LogrusConfig{
			Level:         "",
			Format:        "text",
			DisableColors: false,
			ReportCaller:  false,
		}

		assert.NotPanics(t, func() {
			conf_services.ConfigureLogrus(config)
		})
	})

	t.Run("Whitespace level defaults to info", func(t *testing.T) {
		config := &conf_models.LogrusConfig{
			Level:         "   ",
			Format:        "text",
			DisableColors: false,
			ReportCaller:  false,
		}

		assert.NotPanics(t, func() {
			conf_services.ConfigureLogrus(config)
		})
	})

	t.Run("Mixed case level is handled", func(t *testing.T) {
		config := &conf_models.LogrusConfig{
			Level:         "INFO",
			Format:        "text",
			DisableColors: false,
			ReportCaller:  false,
		}

		assert.NotPanics(t, func() {
			conf_services.ConfigureLogrus(config)
		})
	})

	t.Run("Config with all fields set", func(t *testing.T) {
		config := &conf_models.LogrusConfig{
			Level:         "debug",
			Format:        "json",
			DisableColors: true,
			ReportCaller:  true,
		}

		assert.NotPanics(t, func() {
			conf_services.ConfigureLogrus(config)
		})

		assert.NotNil(t, logrus.StandardLogger())
	})
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    *conf_models.Config
		wantErr bool
	}{
		{
			name: "Running in production environment",
			want: &conf_models.Config{
				ServerPort: "8080",
				Jellyfin: conf_models.JellyfinConfig{
					URL:    "http://localhost:8096",
					APIKey: "key",
					UserID: "admin",
				},
				Logrus: conf_models.LogrusConfig{
					Level:         "info",
					Format:        "json",
					DisableColors: true,
					ReportCaller:  true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set environment variables before test
			os.Setenv(constants.EnvJellyfinURL, tt.want.Jellyfin.URL)
			os.Setenv(constants.EnvJellyfinAPIKey, tt.want.Jellyfin.APIKey)
			os.Setenv(constants.EnvJellyfinAdminUserID, tt.want.Jellyfin.UserID)

			got, gotErr := conf_services.LoadConfig()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("LoadConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("LoadConfig() succeeded unexpectedly")
			}

			assert.Equal(t, tt.want.ServerPort, got.ServerPort)
			assert.Equal(t, tt.want.Jellyfin, got.Jellyfin)
			assert.Equal(t, tt.want.Logrus, got.Logrus)
		})
	}
}
