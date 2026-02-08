package api

import (
	"github.com/typing-master-for-coding-backend/internal/cache"
	"github.com/typing-master-for-coding-backend/internal/config"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/errors"
	"github.com/typing-master-for-coding-backend/internal/handlers"
	"github.com/typing-master-for-coding-backend/internal/logging"
	"github.com/typing-master-for-coding-backend/internal/middleware"
	"github.com/typing-master-for-coding-backend/internal/security"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *database.DatastoreClient, cacheClient *cache.InMemoryCache, cfg *config.Config, logger *logging.Logger, errorHandler *errors.ErrorHandler) *gin.Engine {
	router := gin.Default()

	// Error handling middleware (applied first)
	router.Use(errorHandler.HandlePanic())

	// Cross-platform compatibility middleware
	router.Use(middleware.FeatureDetectionMiddleware())
	router.Use(middleware.AccessibilityMiddleware())
	router.Use(middleware.AccessibilityHeadersMiddleware())
	router.Use(middleware.DeviceCompatibilityMiddleware())
	router.Use(middleware.CompatibilityHeadersMiddleware())
	router.Use(middleware.GracefulDegradationMiddleware())
	router.Use(middleware.FallbackDataMiddleware())

	// Security middleware
	router.Use(middleware.SecurityMiddleware(db))
	router.Use(middleware.RateLimitMiddleware(100))                // 100 requests per minute per IP
	router.Use(middleware.RequestSizeMiddleware(10 * 1024 * 1024)) // 10MB max request size

	// Standard middleware
	router.Use(middleware.CORS(cfg.AllowedOrigins))
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.SecurityAuditMiddleware(db))

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// Compatibility endpoints
	router.GET("/compatibility", handlers.GetCompatibilityInfo)
	router.GET("/compatibility/check", handlers.CheckCompatibility)
	router.GET("/compatibility/fallback", handlers.GetFallbackContent)
	router.GET("/accessibility/settings", handlers.GetAccessibilitySettings)
	router.POST("/compatibility/report", handlers.ReportCompatibilityIssue)

	// Onboarding and help endpoints
	router.GET("/tutorials", handlers.GetTutorials(db))
	router.GET("/tutorials/:id", handlers.GetTutorial(db))
	router.POST("/tutorials/:id/start", handlers.StartTutorial(db))
	router.PUT("/tutorials/:id/progress", handlers.UpdateTutorialProgress(db))
	router.GET("/user/tutorials/progress", handlers.GetUserTutorialProgress(db))
	router.GET("/tooltips", handlers.GetTooltips(db))
	router.POST("/tooltips/:id/viewed", handlers.MarkTooltipViewed(db))
	router.GET("/help/articles/search", handlers.SearchHelpArticles(db))
	router.GET("/help/articles/:id", handlers.GetHelpArticle(db))
	router.POST("/help/articles/:id/rate", handlers.RateHelpArticle(db))
	router.GET("/help/faqs", handlers.GetFAQs(db))
	router.GET("/user/onboarding", handlers.GetOnboardingState(db))
	router.PUT("/user/onboarding/preferences", handlers.UpdateOnboardingPreferences(db))

	// Feedback and support endpoints
	router.POST("/feedback", handlers.CreateFeedback(db))
	router.GET("/user/feedback", handlers.GetUserFeedback(db))
	router.GET("/feedback/categories", handlers.GetFeedbackCategories(db))
	router.POST("/support/tickets", handlers.CreateSupportTicket(db))
	router.GET("/user/support/tickets", handlers.GetSupportTickets(db, ""))
	router.GET("/support/tickets/:id", handlers.GetSupportTicketByID(db))
	router.POST("/support/tickets/:id/messages", handlers.AddSupportMessage(db))
	router.GET("/support/channels", handlers.GetSupportChannels(db))
	router.POST("/support/tickets/:id/rate", handlers.RateSupportTicket(db))

	// Analytics endpoints
	router.POST("/analytics/events", handlers.TrackAnalyticsEvent(db))
	router.POST("/analytics/events/batch", handlers.BatchTrackAnalyticsEvents(db))
	router.GET("/user/analytics/consent", handlers.GetAnalyticsConsent(db))
	router.PUT("/user/analytics/consent", handlers.UpdateAnalyticsConsent(db))

	// Community forum endpoints
	router.GET("/forums", handlers.GetForums(db))
	router.GET("/forums/:id", handlers.GetForum(db))
	router.POST("/forums/:forum_id/posts", handlers.CreateForumPost(db))
	router.GET("/forums/:forum_id/posts", handlers.GetForumPosts(db))
	router.GET("/forum/posts/:id", handlers.GetForumPost(db))
	router.PUT("/forum/posts/:id", handlers.UpdateForumPost(db))
	router.DELETE("/forum/posts/:id", handlers.DeleteForumPost(db))
	router.POST("/forum/posts/:id/replies", handlers.CreateForumReply(db))
	router.GET("/forum/posts/:id/replies", handlers.GetForumReplies(db))
	router.POST("/forum/posts/:id/like", handlers.LikeForumPost(db))
	router.POST("/forum/replies/:reply_id/like", handlers.LikeForumReply(db))
	router.PUT("/forum/posts/:id/answers/:reply_id", handlers.MarkReplyAsAnswer(db))
	router.GET("/user/forum/posts", handlers.GetUserPosts(db))
	router.GET("/forum/posts/search", handlers.SearchForumPosts(db))
	router.GET("/forum/posts/popular", handlers.GetPopularPosts(db))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register(db, cfg.JWTSecret))
			auth.POST("/login", handlers.Login(db, cfg.JWTSecret))
			auth.POST("/login/mfa", handlers.LoginWithMFA(db, cfg.JWTSecret))
			auth.POST("/refresh", handlers.RefreshToken(cfg.JWTSecret))
			auth.POST("/anonymous", handlers.CreateAnonymousSession(cfg.JWTSecret))

			// MFA routes
			auth.POST("/mfa/setup", handlers.SetupMFA(db))
			auth.POST("/mfa/verify-setup", handlers.VerifyMFASetup(db))
			auth.POST("/mfa/disable", handlers.DisableMFA(db))
			auth.POST("/mfa/backup-codes/regenerate", handlers.RegenerateMFABackupCodes(db))
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User routes
			protected.GET("/profile", handlers.GetProfile(db))
			protected.PUT("/profile", handlers.UpdateProfile(db))

			// MFA status (requires authentication)
			protected.GET("/mfa/status", handlers.GetMFAStatus(db))

			// Privacy and GDPR routes
			protected.GET("/privacy/consent", handlers.GetConsentStatus(db))
			protected.PUT("/privacy/settings", handlers.UpdatePrivacySettings(db))
			protected.GET("/privacy/export", handlers.ExportUserData(db))
			protected.POST("/privacy/delete", handlers.DeleteUserData(db))

			// Content routes (read-only for regular users)
			protected.GET("/languages", handlers.GetLanguages(db, cacheClient))
			protected.GET("/languages/:id", handlers.GetLanguage(db, cacheClient))
			protected.GET("/lessons", handlers.GetLessons(db, cacheClient))
			protected.GET("/lessons/:id", handlers.GetLesson(db, cacheClient))
			protected.GET("/snippets", handlers.GetSnippets(db, cacheClient))
			protected.GET("/snippets/:id", handlers.GetSnippet(db, cacheClient))
			protected.GET("/playlists", handlers.GetPlaylists(db, cacheClient))
			protected.GET("/playlists/:id", handlers.GetPlaylist(db, cacheClient))

			// Lesson progression system
			protected.GET("/lessons/:id/progress", handlers.GetLessonProgress(db))
			protected.PUT("/lessons/:id/progress", handlers.UpdateLessonProgress(db))
			protected.GET("/lessons/:id/prerequisites", handlers.CheckLessonPrerequisites(db))
			protected.GET("/lessons/:id/flow", handlers.GetLessonProgressionFlow(db))
			protected.GET("/progression/summary", handlers.GetUserProgressionSummary(db))

			// Maintenance and compliance information
			protected.GET("/system/version", handlers.GetCurrentSystemVersion(db))
			protected.GET("/compliance/standards", handlers.GetActiveComplianceStandards(db))
		}

		// Content creator routes (requires content creation permissions)
		contentCreator := v1.Group("/")
		contentCreator.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		contentCreator.Use(middleware.ContentCreatorAuthMiddleware(db))
		{
			// Content creation and updates (for content creators and above)
			contentCreator.POST("/languages", handlers.CreateLanguage(db))
			contentCreator.PUT("/languages/:id", handlers.UpdateLanguage(db))

			contentCreator.POST("/lessons", handlers.CreateLesson(db))
			contentCreator.PUT("/lessons/:id", handlers.UpdateLesson(db))

			contentCreator.POST("/snippets", handlers.CreateSnippet(db))
			contentCreator.PUT("/snippets/:id", handlers.UpdateSnippet(db))
			contentCreator.POST("/snippets/import", handlers.ImportSnippet(db))

			contentCreator.POST("/playlists", handlers.CreatePlaylist(db))
			contentCreator.PUT("/playlists/:id", handlers.UpdatePlaylist(db))

			// Content versioning and validation
			contentCreator.POST("/content/versions", handlers.CreateContentVersion(db))
			contentCreator.POST("/content/validate", handlers.ValidateContentChecksum(db))
		}

		// Admin-only routes (requires admin role)
		adminOnly := v1.Group("/")
		adminOnly.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		adminOnly.Use(middleware.AdminAuthMiddleware(db))
		{
			// Content deletion (admin only)
			adminOnly.DELETE("/languages/:id", handlers.DeleteLanguage(db))
			adminOnly.DELETE("/lessons/:id", handlers.DeleteLesson(db))
			adminOnly.DELETE("/snippets/:id", handlers.DeleteSnippet(db))
			adminOnly.DELETE("/playlists/:id", handlers.DeletePlaylist(db))

			// User management
			adminOnly.GET("/users", handlers.GetUsers(db))
			adminOnly.GET("/users/:id", handlers.GetUser(db))
			adminOnly.PUT("/users/:id", handlers.UpdateUser(db))
			adminOnly.DELETE("/users/:id", handlers.DeleteUser(db))

			// Analytics and reporting
			adminOnly.GET("/analytics", handlers.GetAnalytics(db))
			adminOnly.GET("/analytics/users", handlers.GetUserAnalytics(db))
			adminOnly.GET("/analytics/content", handlers.GetContentAnalytics(db))

			// Security management routes
			securityLogger := security.NewSecurityLogger(db, security.DefaultLoggerConfig())
			auditService := security.NewAuditService(db, securityLogger)
			dependencyScanner := security.NewDependencyScanner(security.DefaultDependencyScanConfig())

			adminOnly.GET("/security/events", handlers.GetSecurityEvents(db, securityLogger))
			adminOnly.GET("/security/alerts", handlers.GetSecurityAlerts(db, securityLogger))
			adminOnly.PUT("/security/alerts/:id/resolve", handlers.ResolveSecurityAlert(db, securityLogger))
			adminOnly.GET("/security/metrics", handlers.GetSecurityMetrics(db, securityLogger))
			adminOnly.GET("/security/status", handlers.GetSecurityStatus(db, securityLogger, auditService))

			adminOnly.GET("/audit/events", handlers.GetAuditEvents(db, auditService))
			adminOnly.POST("/audit/compliance-report", handlers.GenerateComplianceReport(db, auditService))

			adminOnly.POST("/security/scan-dependencies", handlers.ScanDependencies(dependencyScanner))

			// Monitoring and logging routes (admin only)
			monitoringHandler := handlers.NewMonitoringHandler(db, logger)
			adminOnly.GET("/monitoring/logs", monitoringHandler.GetLogEntries)
			adminOnly.GET("/monitoring/alerts", monitoringHandler.GetErrorAlerts)
			adminOnly.PUT("/monitoring/alerts/:id/resolve", monitoringHandler.ResolveAlert)
			adminOnly.POST("/monitoring/metrics", monitoringHandler.LogPerformanceMetric)
			adminOnly.POST("/monitoring/errors", monitoringHandler.LogError)
			adminOnly.GET("/monitoring/health", monitoringHandler.GetHealthStatus)
			adminOnly.GET("/monitoring/metrics", monitoringHandler.GetMetrics)
			adminOnly.POST("/monitoring/cleanup", monitoringHandler.CleanupLogs)

			// Feedback management (admin only)
			adminOnly.GET("/feedback", handlers.GetFeedback(db))
			adminOnly.GET("/feedback/:id", handlers.GetFeedbackByID(db))
			adminOnly.PUT("/feedback/:id", handlers.UpdateFeedback(db))
			adminOnly.GET("/feedback/stats", handlers.GetFeedbackStats(db))

			// Support ticket management (admin only)
			adminOnly.GET("/support/tickets", handlers.GetSupportTickets(db, ""))
			adminOnly.GET("/support/tickets/:id", handlers.GetSupportTicketByID(db))
			adminOnly.PUT("/support/tickets/:id", handlers.UpdateSupportTicket(db))

			// Analytics management (admin only)
			adminOnly.GET("/analytics/events", handlers.GetAnalyticsEvents(db))
			adminOnly.GET("/analytics/stats", handlers.GetAnalyticsStats(db))

			// A/B Test management (admin only)
			adminOnly.GET("/ab-tests", handlers.GetABTests(db))
			adminOnly.GET("/ab-tests/:id", handlers.GetABTest(db))
			adminOnly.POST("/ab-tests", handlers.CreateABTest(db))
			adminOnly.PUT("/ab-tests/:id", handlers.UpdateABTest(db))
			adminOnly.GET("/ab-tests/:id/results", handlers.GetABTestResults(db))
			adminOnly.GET("/ab-tests/:id/assignments", handlers.GetUserABTestAssignment(db))

			// Community forum management (admin only)
			adminOnly.GET("/forums/stats", handlers.GetForumStats(db))
			adminOnly.POST("/forums", handlers.CreateForum(db))
			adminOnly.PUT("/forums/:id", handlers.UpdateForum(db))
			adminOnly.DELETE("/forums/:id", handlers.DeleteForum(db))

			// System update and compliance management
			adminOnly.GET("/system/versions", handlers.GetSystemVersions(db))
			adminOnly.POST("/system/versions", handlers.CreateSystemVersion(db))
			adminOnly.PUT("/system/versions/:id/current", handlers.SetCurrentSystemVersion(db))

			adminOnly.GET("/compliance/standards/all", handlers.GetComplianceStandards(db))
			adminOnly.POST("/compliance/standards", handlers.CreateComplianceStandard(db))
			adminOnly.PUT("/compliance/standards/:id", handlers.UpdateComplianceStandard(db))
		}

		// Frontend monitoring routes (requires authentication but not admin)
		monitoring := v1.Group("/monitoring")
		monitoring.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			monitoringHandler := handlers.NewMonitoringHandler(db, logger)
			monitoring.POST("/log-error", monitoringHandler.LogError)
			monitoring.POST("/log-errors", monitoringHandler.LogErrors)
			monitoring.POST("/log-metrics", monitoringHandler.LogMetrics)
			monitoring.POST("/finalize-session", monitoringHandler.FinalizeSession)
			monitoring.GET("/health", monitoringHandler.GetHealthStatus)
		}

		// Integration routes (requires authentication)
		integrations := v1.Group("/integrations")
		integrations.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			integrations.POST("/", handlers.CreateIntegration(db))
			integrations.GET("/", handlers.GetIntegrations(db))
			integrations.DELETE("/:id", handlers.DeleteIntegration(db))
			integrations.POST("/github/share", handlers.ShareSnippetToGitHub(db))
			integrations.POST("/embedded/session", handlers.CreateEmbeddedSession(db))
			integrations.GET("/github/repos", handlers.GetGitHubRepos(db))

			// OAuth routes
			integrations.GET("/oauth/:provider", handlers.GetOAuthURL(db))
			integrations.POST("/oauth/:provider/flow", handlers.InitiateOAuthFlow(db))
			integrations.GET("/oauth/:provider/callback", handlers.HandleOAuthCallback(db))
			integrations.POST("/oauth/:provider/refresh", handlers.RefreshOAuthToken(db))
			integrations.DELETE("/oauth/:provider", handlers.RevokeOAuthToken(db))
		}

		// Workflow embedding routes
		embed := v1.Group("/embed")
		{
			embed.POST("/", handlers.CreateEmbed(db))
			embed.GET("/:embed_id", handlers.GetEmbed(db))
			embed.POST("/ci/:platform", handlers.GenerateCIConfig(db))
			embed.GET("/platforms", handlers.GetSupportedPlatforms(db))
			embed.POST("/webhook/:platform", handlers.HandleEmbedWebhook())
		}

		// Public embed routes (no auth required for embedded views)
		public := v1.Group("/public")
		{
			public.GET("/languages", handlers.GetLanguages(db, cacheClient))
			public.GET("/lessons", handlers.GetPublicLessons(db, cacheClient))
			public.GET("/snippets", handlers.GetPublicSnippets(db, cacheClient))
			public.GET("/embed/session/:session_id", handlers.GetEmbedSession(db))
		}
	}

	return router
}
