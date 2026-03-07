package services

import (
	"encoding/json"

	conf_models "jellyfin-duplicate/configuration/models"
	"jellyfin-duplicate/configuration/services/assets"
	"jellyfin-duplicate/constants"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func loadEnv() conf_models.Config {
	if err := godotenv.Load(); err != nil {
		logrus.Infof("No .env file loaded or error reading it: %v", err)
	}

	// Check required environment variables
	requiredVars := []string{constants.EnvJellyfinURL, constants.EnvJellyfinAPIKey, constants.EnvJellyfinAdminUserID}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			logrus.Fatalf("Environment variable %s not set", v)
		}
	}

	logrus.Infof("Running in %s environment", assets.GetEnv())

	return conf_models.Config{
		Jellyfin: conf_models.JellyfinConfig{
			URL:    os.Getenv(constants.EnvJellyfinURL),
			APIKey: os.Getenv(constants.EnvJellyfinAPIKey),
			UserID: os.Getenv(constants.EnvJellyfinAdminUserID),
		},
	}
}

func LoadConfig() (*conf_models.Config, error) {

	// Load environment variables from .env file
	config := loadEnv()

	// Read config file
	file, err := assets.GetConfigFile()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	// Merge config with environment variables and config file
	return &config, nil
}

func ConfigureLogrus(config *conf_models.LogrusConfig) {
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		logrus.Warnf("Invalid log level '%s', defaulting to Info", config.Level)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Set formatter based on format configuration
	if config.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: false,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors: config.DisableColors,
			FullTimestamp: true,
		})
	}

	// Set report caller
	logrus.SetReportCaller(config.ReportCaller)
}

func ConfigureGINMode() {
	if assets.GetEnv() == constants.Production {
		gin.SetMode(gin.ReleaseMode)
		logrus.Info("GIN set to Release mode")
	} else {
		logrus.Info("GIN set to Debug mode")
	}
}
