package api

import (
	"typing-master-backend/internal/cache"
	"typing-master-backend/internal/config"
	"typing-master-backend/internal/database"
	"typing-master-backend/internal/handlers"
	"typing-master-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *database.DatastoreClient, cacheClient *cache.InMemoryCache, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS(cfg.AllowedOrigins))
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register(db))
			auth.POST("/login", handlers.Login(db, cfg.JWTSecret))
			auth.POST("/refresh", handlers.RefreshToken(cfg.JWTSecret))
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User routes
			protected.GET("/profile", handlers.GetProfile(db))
			protected.PUT("/profile", handlers.UpdateProfile(db))

			// Content routes
			protected.GET("/languages", handlers.GetLanguages(db, cacheClient))
			protected.GET("/languages/:id", handlers.GetLanguage(db, cacheClient))
			protected.POST("/languages", handlers.CreateLanguage(db))
			protected.PUT("/languages/:id", handlers.UpdateLanguage(db))
			protected.DELETE("/languages/:id", handlers.DeleteLanguage(db))

			protected.GET("/lessons", handlers.GetLessons(db, cacheClient))
			protected.GET("/lessons/:id", handlers.GetLesson(db, cacheClient))
			protected.POST("/lessons", handlers.CreateLesson(db))
			protected.PUT("/lessons/:id", handlers.UpdateLesson(db))
			protected.DELETE("/lessons/:id", handlers.DeleteLesson(db))

			protected.GET("/snippets", handlers.GetSnippets(db, cacheClient))
			protected.GET("/snippets/:id", handlers.GetSnippet(db, cacheClient))
			protected.POST("/snippets", handlers.CreateSnippet(db))
			protected.PUT("/snippets/:id", handlers.UpdateSnippet(db))
			protected.DELETE("/snippets/:id", handlers.DeleteSnippet(db))
			protected.POST("/snippets/import", handlers.ImportSnippet(db))

			protected.GET("/playlists", handlers.GetPlaylists(db, cacheClient))
			protected.GET("/playlists/:id", handlers.GetPlaylist(db, cacheClient))
			protected.POST("/playlists", handlers.CreatePlaylist(db))
			protected.PUT("/playlists/:id", handlers.UpdatePlaylist(db))
			protected.DELETE("/playlists/:id", handlers.DeletePlaylist(db))

			// Content versioning and validation
			protected.POST("/content/versions", handlers.CreateContentVersion(db))
			protected.GET("/content/versions/:contentType/:contentId", handlers.GetContentVersions(db))
			protected.POST("/content/validate", handlers.ValidateContentChecksum(db))

			// Lesson progression system
			protected.GET("/lessons/:id/progress", handlers.GetLessonProgress(db))
			protected.PUT("/lessons/:id/progress", handlers.UpdateLessonProgress(db))
			protected.GET("/lessons/:id/prerequisites", handlers.CheckLessonPrerequisites(db))
			protected.GET("/lessons/:id/flow", handlers.GetLessonProgressionFlow(db))
			protected.GET("/progression/summary", handlers.GetUserProgressionSummary(db))

			// Session routes
			protected.POST("/sessions", handlers.CreateSession(db, cacheClient))
			protected.PUT("/sessions/:id", handlers.UpdateSession(db, cacheClient))
			protected.POST("/sessions/:id/events", handlers.RecordEvents(db, cacheClient))
			protected.POST("/sessions/:id/finalize", handlers.FinalizeSession(db, cacheClient))

			// Results routes
			protected.GET("/results", handlers.GetResults(db))
			protected.GET("/results/:session_id", handlers.GetResult(db))

			// Leaderboard routes
			protected.GET("/leaderboards", handlers.GetLeaderboards(cacheClient))
		}

		// Public routes (no auth required)
		public := v1.Group("/public")
		{
			public.GET("/languages", handlers.GetLanguages(db, cacheClient))
			public.GET("/lessons", handlers.GetPublicLessons(db, cacheClient))
			public.GET("/snippets", handlers.GetPublicSnippets(db, cacheClient))
		}
	}

	return router
}