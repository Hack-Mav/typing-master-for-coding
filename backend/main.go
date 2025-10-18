package main

import (
	"log"
	"os"

	"typing-master-backend/internal/api"
	"typing-master-backend/internal/cache"
	"typing-master-backend/internal/config"
	"typing-master-backend/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize Datastore
	db, err := database.Initialize(cfg.ProjectID)
	if err != nil {
		log.Fatal("Failed to initialize datastore:", err)
	}
	defer db.Close()

	// Initialize in-memory cache
	cacheClient := cache.NewInMemoryCache(cfg.CacheMaxSize, cfg.CacheTTLMinutes)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := api.SetupRouter(db, cacheClient, cfg)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Project ID: %s", cfg.ProjectID)
	log.Printf("Environment: %s", cfg.Environment)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}