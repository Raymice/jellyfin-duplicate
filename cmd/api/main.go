package main

import (
	"encoding/json"
	"io/ioutil"
	"jellyfin-duplicate/internal/handlers"
	"jellyfin-duplicate/internal/jellyfin"
	"jellyfin-duplicate/internal/models"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type LogrusConfig struct {
	Level         string `json:"level"`
	Format        string `json:"format"`
	DisableColors bool   `json:"disable_colors"`
	ReportCaller  bool   `json:"report_caller"`
}

type Config struct {
	JellyfinURL string       `json:"jellyfin_url"`
	APIKey      string       `json:"api_key"`
	UserID      string       `json:"user_id"`
	ServerPort  string       `json:"server_port"`
	Logrus      LogrusConfig `json:"logrus"`
}

func configureLogrus(config *LogrusConfig) {
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

func getConfigPath(environment models.Environment) (path string) {

	switch environment {
	case models.Development:

		if _, err := os.Stat("config.dev.json"); err == nil {
			path = "config.dev.json"
			return
		} else {
			logrus.Fatalf("config.dev.json file not found")
		}

	case models.Production:
		if _, err := os.Stat("config.prod.json"); err == nil {
			path = "config.prod.json"
			return
		} else {
			logrus.Fatalf("config.prod.json file not found")
		}
	default:
		logrus.Fatalf("Invalid environment variable: %s", environment)
	}

	return
}

func getEnvironment() models.Environment {
	// Default
	environment := models.Development

	env := os.Getenv("ENVIRONMENT")

	if env == "" {
		logrus.Warn("ENVIRONMENT variable not set, defaulting to 'development'")
		environment = models.Development
	} else {
		environment = models.Environment(env)
	}

	if environment != models.Development && environment != models.Production {
		logrus.Fatalf("Invalid ENVIRONMENT value: %s. Must be 'development' or 'production'", env)
	}

	logrus.Infof("Running in %s environment", environment)

	return environment
}

func loadConfig(environment models.Environment) (*Config, error) {
	configPath := getConfigPath(environment)
	logrus.Infof("Loading configuration from: %s", configPath)

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	// Initialize with default logrus settings first
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Info("Starting jellyfin-duplicate application...")

	// Determine environment
	environment := getEnvironment()

	// Load configuration
	logrus.Info("Loading configuration...")
	config, err := loadConfig(environment)
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// Configure logrus based on config
	configureLogrus(&config.Logrus)

	logrus.Infof("Configuration loaded successfully. Jellyfin URL: %s", config.JellyfinURL)

	// Configure GIN mode
	if environment == models.Production {
		gin.SetMode(gin.ReleaseMode)
		logrus.Info("GIN set to Release mode")
	} else {
		logrus.Info("GIN set to Debug mode")
	}

	// Initialize Jellyfin client
	logrus.Info("Initializing Jellyfin client...")
	jellyfinClient := jellyfin.NewClient(config.JellyfinURL, config.APIKey)

	// Set user ID for library access
	logrus.Infof("Setting user ID: %s", config.UserID)
	if err := jellyfinClient.SetUserID(config.UserID); err != nil {
		logrus.Fatalf("Failed to set user ID: %v", err)
	}
	logrus.Info("Jellyfin client initialized successfully")

	// Create Gin router
	logrus.Info("Setting up web server...")
	r := gin.Default()

	// Load HTML templates
	logrus.Info("Loading HTML templates...")
	r.LoadHTMLGlob("web/templates/*")

	// Set up handlers
	logrus.Info("Initializing handlers...")
	handler := handlers.NewHandler(jellyfinClient)

	// Routes
	logrus.Info("Configuring routes...")
	r.GET("/", handler.GetDuplicatesPage)
	r.GET("/api/duplicates", handler.GetDuplicatesJSON)
	r.GET("/api/mark-as-seen", handler.MarkMovieAsSeen)
	r.GET("/api/delete-movie", handler.DeleteMovie)
	logrus.Info("Routes configured successfully")

	// Start server
	port := ":" + config.ServerPort
	logrus.Infof("Starting server on %s", port)
	logrus.Infof("Application ready. Access the web interface at http://localhost%s", port)
	if err := r.Run(port); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
