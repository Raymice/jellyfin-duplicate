package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"jellyfin-duplicate/internal/handlers"
	"jellyfin-duplicate/internal/jellyfin"
	"github.com/gin-gonic/gin"
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
	log.Println("Starting jellyfin-duplicate application...")
	
	// Load configuration
	log.Println("Loading configuration...")
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Configuration loaded successfully. Jellyfin URL: %s", config.JellyfinURL)
	
	// Initialize Jellyfin client
	log.Println("Initializing Jellyfin client...")
	jellyfinClient := jellyfin.NewClient(config.JellyfinURL, config.APIKey)
	
	// Set user ID for library access
	log.Printf("Setting user ID: %s", config.UserID)
	if err := jellyfinClient.SetUserID(config.UserID); err != nil {
		log.Fatalf("Failed to set user ID: %v", err)
	}
	log.Println("Jellyfin client initialized successfully")

	// Create Gin router
	log.Println("Setting up web server...")
	r := gin.Default()

	// Load HTML templates
	log.Println("Loading HTML templates...")
	r.LoadHTMLGlob("web/templates/*")

	// Set up handlers
	log.Println("Initializing handlers...")
	handler := handlers.NewHandler(jellyfinClient)

	// Routes
	log.Println("Configuring routes...")
	r.GET("/", handler.GetDuplicatesPage)
	r.GET("/api/duplicates", handler.GetDuplicatesJSON)
	r.GET("/api/mark-as-seen", handler.MarkMovieAsSeen)
	r.GET("/api/delete-movie", handler.DeleteMovie)
	log.Println("Routes configured successfully")

	// Start server
	port := ":" + config.ServerPort
	log.Printf("Starting server on %s", port)
	log.Printf("Application ready. Access the web interface at http://localhost%s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}