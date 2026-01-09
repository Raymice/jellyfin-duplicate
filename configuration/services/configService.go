package services

import (
	"encoding/json"
	conf_models "jellyfin-duplicate/configuration/models"
	"jellyfin-duplicate/constants"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func getConfigPath(environment constants.Environment) (path string) {

	configurationDir := "configuration/files"
	prodFile := "config.prod.json"
	devFile := "config.dev.json"

	switch environment {
	case constants.Development:

		if _, err := os.Stat(configurationDir + "/" + devFile); err == nil {
			path = configurationDir + "/" + devFile
			return
		} else {
			logrus.Fatalf("%s file not found", devFile)
		}

	case constants.Production:
		if _, err := os.Stat(configurationDir + "/" + prodFile); err == nil {
			path = configurationDir + "/" + prodFile
			return
		} else {
			logrus.Fatalf("%s file not found", prodFile)
		}
	default:
		logrus.Fatalf("Invalid environment variable: %s", environment)
	}

	return
}

func loadEnv() conf_models.Config {
	if err := godotenv.Load(); err != nil {
		logrus.Infof("No .env file loaded or error reading it: %v", err)
	}

	// Check required environment variables
	requiredVars := []string{constants.EnvJellyfinURL, constants.EnvJellyfinAPIKey, constants.EnvJellyfinAdminUserID, constants.EnvEnvironment}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			logrus.Fatalf("Environment variable %s not set", v)
		}
	}

	env := os.Getenv(constants.EnvEnvironment)
	if env != string(constants.Development) && env != string(constants.Production) {
		logrus.Fatalf("Invalid ENVIRONMENT value: %s. Must be 'development' or 'production'", env)
	}

	logrus.Infof("Running in %s environment", env)

	return conf_models.Config{
		Environment: constants.Environment(env),
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

	configPath := getConfigPath(config.Environment)
	logrus.Infof("Loading configuration from: %s", configPath)

	// Read config file
	file, err := os.ReadFile(configPath)
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

func ConfigureGINMode(environment constants.Environment) {
	if environment == constants.Production {
		gin.SetMode(gin.ReleaseMode)
		logrus.Info("GIN set to Release mode")
	} else {
		logrus.Info("GIN set to Debug mode")
	}
}
