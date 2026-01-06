package main

import (
	jellyfinClient "jellyfin-duplicate/client/jellyfin/http"
	confServices "jellyfin-duplicate/configuration/services"
	server "jellyfin-duplicate/server"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize with default logrus settings first
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Info("Starting jellyfin-duplicate application...")

	// Load configuration
	logrus.Info("Loading configuration...")
	config, err := confServices.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// Configure logrus based on config
	confServices.ConfigureLogrus(&config.Logrus)

	logrus.Infof("Configuration loaded successfully. Jellyfin URL: %s", config.Jellyfin.URL)

	// Configure GIN mode
	confServices.ConfigureGINMode(config.Environment)

	// Initialize Jellyfin client
	logrus.Info("Initializing Jellyfin client...")
	jellyfinClient := jellyfinClient.NewClient(config.Jellyfin.URL, config.Jellyfin.APIKey, config.Jellyfin.UserID)

	logrus.Info("Jellyfin client initialized successfully")

	// Create Gin router
	logrus.Info("Setting up web server...")
	r := gin.Default()

	// Load HTML templates
	logrus.Info("Loading HTML templates...")
	r.LoadHTMLGlob("server/templates/*")

	// Set up handlers
	logrus.Info("Initializing handlers...")
	handler := server.NewHandler(jellyfinClient)

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
