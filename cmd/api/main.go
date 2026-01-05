package main

import (
	"encoding/json"
	"io/ioutil"
	"jellyfin-duplicate/internal/handlers"
	"jellyfin-duplicate/internal/jellyfin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Config struct {
	JellyfinURL string `json:"jellyfin_url"`
	APIKey      string `json:"api_key"`
	UserID      string `json:"user_id"`
	ServerPort  string `json:"server_port"`
}

func loadConfig() (*Config, error) {
	file, err := ioutil.ReadFile("config.json")
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
	logrus.Info("Starting jellyfin-duplicate application...")
	
	// Load configuration
	logrus.Info("Loading configuration...")
	config, err := loadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}
	logrus.Infof("Configuration loaded successfully. Jellyfin URL: %s", config.JellyfinURL)
	
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