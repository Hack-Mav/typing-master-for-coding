package main

import (
	"log"
	"os"

	"github.com/typing-master-for-coding-backend/internal/api"
	"github.com/typing-master-for-coding-backend/internal/cache"
	"github.com/typing-master-for-coding-backend/internal/config"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/errors"
	"github.com/typing-master-for-coding-backend/internal/handlers"
	"github.com/typing-master-for-coding-backend/internal/logging"

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

	// Initialize database (Postgres or in-memory mock)
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize cache (Redis when configured, otherwise in-memory)
	cacheClient := cache.NewCache(cfg)

	// Initialize default languages
	if err := database.InitializeDefaultLanguages(db); err != nil {
		log.Printf("Warning: Failed to initialize default languages: %v", err)
	} else {
		log.Println("Default languages initialized successfully")
	}

	// Initialize default assessments
	if err := database.InitializeDefaultAssessments(db); err != nil {
		log.Printf("Warning: Failed to initialize default assessments: %v", err)
	} else {
		log.Println("Default assessments initialized successfully")
	}

	// Initialize scoring services
	handlers.InitializeScoringServices(db)
	log.Println("Scoring services initialized successfully")

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize logger and error handler
	logger := logging.NewLogger(db, logging.DefaultLoggerConfig())
	errorHandler := errors.NewErrorHandler(logger, errors.DefaultErrorHandlerConfig())

	// Initialize router
	router := api.SetupRouter(db, cacheClient, cfg, logger, errorHandler)

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
