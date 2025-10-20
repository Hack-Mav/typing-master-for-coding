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
			auth.POST("/register", handlers.Register(db, cfg.JWTSecret))
			auth.POST("/login", handlers.Login(db, cfg.JWTSecret))
			auth.POST("/refresh", handlers.RefreshToken(cfg.JWTSecret))
			auth.POST("/anonymous", handlers.CreateAnonymousSession(cfg.JWTSecret))
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User routes
			protected.GET("/profile", handlers.GetProfile(db))
			protected.PUT("/profile", handlers.UpdateProfile(db))

			// Privacy and GDPR routes
			protected.GET("/privacy/consent", handlers.GetConsentStatus(db))
			protected.PUT("/privacy/settings", handlers.UpdatePrivacySettings(db))
			protected.GET("/privacy/export", handlers.ExportUserData(db))
			protected.POST("/privacy/delete", handlers.DeleteUserData(db))

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
			protected.GET("/leaderboards/rank/:user_id", handlers.GetUserRank)
			protected.GET("/leaderboards/trends", handlers.GetLeaderboardTrends)

			// Scoring and metrics routes
			protected.POST("/scoring/process", handlers.ProcessSessionMetrics)
			protected.GET("/scoring/anticheat/:session_id", handlers.AnalyzeAntiCheat)

			// Tournament routes
			protected.POST("/tournaments", handlers.CreateTournament)
			protected.GET("/tournaments/:tournament_id", handlers.GetTournament)
			protected.POST("/tournaments/:tournament_id/register", handlers.RegisterForTournament)
			protected.GET("/tournaments/:tournament_id/leaderboard", handlers.GetTournamentLeaderboard)
			protected.POST("/tournaments/:tournament_id/submit", handlers.SubmitTournamentResult)
			protected.GET("/tournaments/user/:user_id", handlers.GetUserTournaments)

			// Assessment routes
			protected.GET("/assessments/blueprints", handlers.GetAssessmentBlueprints(db, cacheClient))
			protected.POST("/assessments/blueprints", handlers.CreateAssessmentBlueprint(db, cacheClient))
			protected.GET("/assessments/blueprints/:id/snippets/:snippetId", handlers.GetAssessmentSnippet(db, cacheClient))
			protected.GET("/assessments/blueprints/:id", handlers.GetAssessmentBlueprint(db, cacheClient))
			protected.POST("/assessments/sessions", handlers.CreateAssessmentSession(db, cacheClient))
			protected.GET("/assessments/sessions/:id", handlers.GetAssessmentSession(db, cacheClient))
			protected.POST("/assessments/sessions/:id/snippets", handlers.RecordSnippetResult(db, cacheClient))
			protected.POST("/assessments/sessions/:id/finalize", handlers.FinalizeAssessment(db, cacheClient))
			protected.GET("/assessments/analytics", handlers.GetAssessmentAnalytics(db, cacheClient))
			protected.GET("/assessments/badges", handlers.GetUserBadges(db, cacheClient))
			protected.POST("/assessments/schedule", handlers.ScheduleAssessment(db, cacheClient))
		}

		// Admin routes (requires admin privileges)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		admin.Use(middleware.AdminAuthMiddleware())
		{
			// Dashboard
			admin.GET("/dashboard", handlers.GetAdminDashboard(db))

			// Content management
			admin.GET("/languages", handlers.GetAdminLanguages(db))
			admin.POST("/languages", handlers.CreateAdminLanguage(db))
			admin.PUT("/languages/:id", handlers.UpdateAdminLanguage(db))
			admin.DELETE("/languages/:id", handlers.DeleteAdminLanguage(db))

			admin.GET("/lessons", handlers.GetAdminLessons(db))
			admin.POST("/lessons", handlers.CreateAdminLesson(db))
			admin.PUT("/lessons/:id", handlers.UpdateAdminLesson(db))
			admin.DELETE("/lessons/:id", handlers.DeleteAdminLesson(db))

			admin.GET("/snippets", handlers.GetAdminSnippets(db))
			admin.POST("/snippets", handlers.CreateAdminSnippet(db))
			admin.PUT("/snippets/:id", handlers.UpdateAdminSnippet(db))
			admin.DELETE("/snippets/:id", handlers.DeleteAdminSnippet(db))

			// Content versioning
			admin.GET("/content/versions/:contentType/:contentId", handlers.GetAdminContentVersions(db))
			admin.POST("/content/versions/:contentType/:contentId/restore/:version", handlers.RestoreAdminContentVersion(db))
			admin.POST("/content/validate", handlers.ValidateAdminContent(db))

			// A/B Testing
			admin.GET("/ab-tests", handlers.GetABTests(db))
			admin.POST("/ab-tests", handlers.CreateABTest(db))
			admin.PUT("/ab-tests/:id", handlers.UpdateABTest(db))
			admin.DELETE("/ab-tests/:id", handlers.DeleteABTest(db))
			admin.GET("/ab-tests/:id/results", handlers.GetABTestResults(db))

			// Tournament management
			admin.POST("/tournaments/:tournament_id/start", handlers.StartTournament)
			admin.POST("/tournaments/:tournament_id/end", handlers.EndTournament)
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