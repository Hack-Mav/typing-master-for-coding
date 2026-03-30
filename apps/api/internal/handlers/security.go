package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/security"

	"github.com/gin-gonic/gin"
)

// GetSecurityEvents retrieves security events with filtering
func GetSecurityEvents(db *database.DatastoreClient, logger *security.SecurityLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		filters := security.SecurityEventFilters{}

		if eventType := c.Query("type"); eventType != "" {
			filters.Type = eventType
		}

		if severity := c.Query("severity"); severity != "" {
			filters.Severity = severity
		}

		if ip := c.Query("ip"); ip != "" {
			filters.IP = ip
		}

		if startTime := c.Query("start_time"); startTime != "" {
			if t, err := time.Parse(time.RFC3339, startTime); err == nil {
				filters.StartTime = t
			}
		}

		if endTime := c.Query("end_time"); endTime != "" {
			if t, err := time.Parse(time.RFC3339, endTime); err == nil {
				filters.EndTime = t
			}
		}

		if limit := c.Query("limit"); limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				filters.Limit = l
			}
		}

		// Get security events
		events, err := logger.GetSecurityEvents(c.Request.Context(), filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve security events"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"events": events,
			"count":  len(events),
		})
	}
}

// GetSecurityAlerts retrieves security alerts
func GetSecurityAlerts(db *database.DatastoreClient, logger *security.SecurityLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		includeResolved := c.Query("include_resolved") == "true"

		alerts, err := logger.GetSecurityAlerts(c.Request.Context(), includeResolved)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve security alerts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"alerts": alerts,
			"count":  len(alerts),
		})
	}
}

// ResolveSecurityAlert marks a security alert as resolved
func ResolveSecurityAlert(db *database.DatastoreClient, logger *security.SecurityLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		alertID := c.Param("id")
		if alertID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID is required"})
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		err := logger.ResolveAlert(c.Request.Context(), alertID, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve alert"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Alert resolved successfully"})
	}
}

// GetSecurityMetrics returns security metrics for monitoring
func GetSecurityMetrics(db *database.DatastoreClient, logger *security.SecurityLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse window parameter (default to 24 hours)
		windowStr := c.DefaultQuery("window", "24h")
		window, err := time.ParseDuration(windowStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid window duration"})
			return
		}

		metrics, err := logger.GetSecurityMetrics(c.Request.Context(), window)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve security metrics"})
			return
		}

		c.JSON(http.StatusOK, metrics)
	}
}

// GetAuditEvents retrieves audit events with filtering
func GetAuditEvents(db *database.DatastoreClient, auditService *security.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		filters := security.AuditEventFilters{}

		if eventType := c.Query("type"); eventType != "" {
			filters.Type = eventType
		}

		if action := c.Query("action"); action != "" {
			filters.Action = action
		}

		if resource := c.Query("resource"); resource != "" {
			filters.Resource = resource
		}

		if userID := c.Query("user_id"); userID != "" {
			filters.UserID = userID
		}

		if startTime := c.Query("start_time"); startTime != "" {
			if t, err := time.Parse(time.RFC3339, startTime); err == nil {
				filters.StartTime = t
			}
		}

		if endTime := c.Query("end_time"); endTime != "" {
			if t, err := time.Parse(time.RFC3339, endTime); err == nil {
				filters.EndTime = t
			}
		}

		if limit := c.Query("limit"); limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				filters.Limit = l
			}
		}

		// Get audit events
		events, err := auditService.GetAuditEvents(c.Request.Context(), filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit events"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"events": events,
			"count":  len(events),
		})
	}
}

// GenerateComplianceReport generates a compliance audit report
func GenerateComplianceReport(db *database.DatastoreClient, auditService *security.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			StartDate  string `json:"start_date" binding:"required"`
			EndDate    string `json:"end_date" binding:"required"`
			ReportType string `json:"report_type"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
			return
		}

		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
			return
		}

		// Set end date to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		if req.ReportType == "" {
			req.ReportType = "general"
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		report, err := auditService.GenerateComplianceReport(
			c.Request.Context(),
			startDate,
			endDate,
			req.ReportType,
			userID.(string),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate compliance report"})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// ScanDependencies performs a dependency vulnerability scan
func ScanDependencies(scanner *security.DependencyScanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ProjectPath string `json:"project_path"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.ProjectPath == "" {
			req.ProjectPath = "." // Default to current directory
		}

		result, err := scanner.ScanDependencies(c.Request.Context(), req.ProjectPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan dependencies",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetSecurityStatus returns overall security status
func GetSecurityStatus(db *database.DatastoreClient, logger *security.SecurityLogger, auditService *security.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get recent security metrics (last 24 hours)
		metrics, err := logger.GetSecurityMetrics(ctx, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve security metrics"})
			return
		}

		// Get unresolved alerts
		alerts, err := logger.GetSecurityAlerts(ctx, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve security alerts"})
			return
		}

		// Calculate security score based on recent activity
		securityScore := calculateSecurityScore(metrics, alerts)

		status := gin.H{
			"security_score":     securityScore,
			"total_events":       metrics.TotalEvents,
			"blocked_events":     metrics.BlockedEvents,
			"active_alerts":      len(alerts),
			"events_by_type":     metrics.EventsByType,
			"events_by_severity": metrics.EventsBySeverity,
			"last_updated":       time.Now().UTC(),
		}

		c.JSON(http.StatusOK, status)
	}
}

// calculateSecurityScore calculates a security score based on metrics and alerts
func calculateSecurityScore(metrics *security.SecurityMetrics, alerts []security.SecurityAlert) float64 {
	baseScore := 100.0

	// Deduct points for security events
	if metrics.TotalEvents > 100 {
		baseScore -= 20.0
	} else if metrics.TotalEvents > 50 {
		baseScore -= 10.0
	} else if metrics.TotalEvents > 20 {
		baseScore -= 5.0
	}

	// Deduct points for active alerts
	for _, alert := range alerts {
		switch alert.Severity {
		case "CRITICAL":
			baseScore -= 15.0
		case "HIGH":
			baseScore -= 10.0
		case "MEDIUM":
			baseScore -= 5.0
		case "LOW":
			baseScore -= 2.0
		}
	}

	// Bonus points for having blocked events (shows security is working)
	if metrics.BlockedEvents > 0 && metrics.TotalEvents > 0 {
		blockRate := float64(metrics.BlockedEvents) / float64(metrics.TotalEvents)
		if blockRate > 0.8 {
			baseScore += 5.0
		}
	}

	// Ensure score doesn't go below 0 or above 100
	if baseScore < 0 {
		baseScore = 0
	} else if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}
