package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/logging"

	"github.com/gin-gonic/gin"
)

// MonitoringHandler handles monitoring and logging endpoints
type MonitoringHandler struct {
	db     *database.DatastoreClient
	logger *logging.Logger
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(db *database.DatastoreClient, logger *logging.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		db:     db,
		logger: logger,
	}
}

// GetLogEntries retrieves log entries with filtering
func (h *MonitoringHandler) GetLogEntries(c *gin.Context) {
	// Parse query parameters
	filters := logging.LogFilters{
		Level:     c.Query("level"),
		Service:   c.Query("service"),
		Component: c.Query("component"),
		UserID:    c.Query("user_id"),
		SessionID: c.Query("session_id"),
	}

	// Parse time filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filters.StartTime = startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filters.EndTime = endTime
		}
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	// Get log entries
	entries, err := h.logger.GetLogEntries(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to get log entries", err, map[string]interface{}{
			"filters": filters,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve log entries",
		})
		return
	}

	h.logger.Info("Log entries retrieved", map[string]interface{}{
		"count":   len(entries),
		"filters": filters,
	})

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"count":   len(entries),
		"filters": filters,
	})
}

// GetErrorAlerts retrieves error alerts
func (h *MonitoringHandler) GetErrorAlerts(c *gin.Context) {
	includeResolved := c.Query("include_resolved") == "true"

	alerts, err := h.logger.GetErrorAlerts(c.Request.Context(), includeResolved)
	if err != nil {
		h.logger.Error("Failed to get error alerts", err, map[string]interface{}{
			"include_resolved": includeResolved,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve error alerts",
		})
		return
	}

	h.logger.Info("Error alerts retrieved", map[string]interface{}{
		"count":            len(alerts),
		"include_resolved": includeResolved,
	})

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// ResolveAlert marks an error alert as resolved
func (h *MonitoringHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	var request struct {
		ResolvedBy string `json:"resolved_by" binding:"required"`
		Notes      string `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := h.logger.ResolveAlert(c.Request.Context(), alertID, request.ResolvedBy)
	if err != nil {
		h.logger.Error("Failed to resolve alert", err, map[string]interface{}{
			"alert_id":    alertID,
			"resolved_by": request.ResolvedBy,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resolve alert",
		})
		return
	}

	h.logger.Info("Alert resolved", map[string]interface{}{
		"alert_id":    alertID,
		"resolved_by": request.ResolvedBy,
		"notes":       request.Notes,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert resolved successfully",
	})
}

// LogPerformanceMetric logs a performance metric
func (h *MonitoringHandler) LogPerformanceMetric(c *gin.Context) {
	var request struct {
		Name      string                 `json:"name" binding:"required"`
		Value     float64                `json:"value" binding:"required"`
		Unit      string                 `json:"unit"`
		Tags      map[string]interface{} `json:"tags"`
		Timestamp *time.Time             `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Use current time if not provided
	if request.Timestamp == nil {
		now := time.Now().UTC()
		request.Timestamp = &now
	}

	h.logger.LogPerformanceMetric(
		c.Request.Context(),
		request.Name,
		request.Value,
		request.Unit,
		request.Tags,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Performance metric logged successfully",
	})
}

// LogError logs an error with user-friendly message
func (h *MonitoringHandler) LogError(c *gin.Context) {
	var request struct {
		Error struct {
			ID              string                 `json:"id"`
			Type            string                 `json:"type"`
			Severity        string                 `json:"severity"`
			Message         string                 `json:"message"`
			UserMessage     string                 `json:"user_message"`
			Timestamp       int64                  `json:"timestamp"`
			Retryable       bool                   `json:"retryable"`
			SuggestedAction string                 `json:"suggested_action,omitempty"`
			Context         map[string]interface{} `json:"context,omitempty"`
		} `json:"error" binding:"required"`
		Session struct {
			SessionID string `json:"sessionId"`
			UserID    string `json:"userId,omitempty"`
		} `json:"session,omitempty"`
		Immediate bool `json:"immediate,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Create metadata from error context and session
	metadata := make(map[string]interface{})
	if request.Error.Context != nil {
		for k, v := range request.Error.Context {
			metadata[k] = v
		}
	}
	metadata["error_id"] = request.Error.ID
	metadata["error_type"] = request.Error.Type
	metadata["severity"] = request.Error.Severity
	metadata["retryable"] = request.Error.Retryable
	metadata["suggested_action"] = request.Error.SuggestedAction
	metadata["session_id"] = request.Session.SessionID
	metadata["user_id"] = request.Session.UserID
	metadata["immediate"] = request.Immediate

	// Log based on severity
	switch request.Error.Severity {
	case "CRITICAL":
		h.logger.Error(request.Error.Message, nil, metadata)
	case "HIGH":
		h.logger.Error(request.Error.Message, nil, metadata)
	case "MEDIUM":
		h.logger.Warn(request.Error.Message, metadata)
	default:
		h.logger.Info(request.Error.Message, metadata)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Error logged successfully",
		"error_id": request.Error.ID,
	})
}

// LogErrors logs multiple errors in batch
func (h *MonitoringHandler) LogErrors(c *gin.Context) {
	var request struct {
		Errors []struct {
			ID              string                 `json:"id"`
			Type            string                 `json:"type"`
			Severity        string                 `json:"severity"`
			Message         string                 `json:"message"`
			UserMessage     string                 `json:"user_message"`
			Timestamp       int64                  `json:"timestamp"`
			Retryable       bool                   `json:"retryable"`
			SuggestedAction string                 `json:"suggested_action,omitempty"`
			Context         map[string]interface{} `json:"context,omitempty"`
		} `json:"errors" binding:"required"`
		Session struct {
			SessionID string `json:"sessionId"`
			UserID    string `json:"userId,omitempty"`
		} `json:"session,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	errorIDs := make([]string, len(request.Errors))

	for i, errorData := range request.Errors {
		// Create metadata for each error
		metadata := make(map[string]interface{})
		if errorData.Context != nil {
			for k, v := range errorData.Context {
				metadata[k] = v
			}
		}
		metadata["error_id"] = errorData.ID
		metadata["error_type"] = errorData.Type
		metadata["severity"] = errorData.Severity
		metadata["retryable"] = errorData.Retryable
		metadata["suggested_action"] = errorData.SuggestedAction
		metadata["session_id"] = request.Session.SessionID
		metadata["user_id"] = request.Session.UserID
		metadata["batch_index"] = i

		// Log based on severity
		switch errorData.Severity {
		case "CRITICAL":
			h.logger.Error(errorData.Message, nil, metadata)
		case "HIGH":
			h.logger.Error(errorData.Message, nil, metadata)
		case "MEDIUM":
			h.logger.Warn(errorData.Message, metadata)
		default:
			h.logger.Info(errorData.Message, metadata)
		}

		errorIDs[i] = errorData.ID
	}

	h.logger.Info("Batch error logging completed", map[string]interface{}{
		"batch_size": len(request.Errors),
		"session_id": request.Session.SessionID,
		"user_id":    request.Session.UserID,
		"error_ids":  errorIDs,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Errors logged successfully",
		"count":     len(request.Errors),
		"error_ids": errorIDs,
	})
}

// LogMetrics logs performance metrics in batch
func (h *MonitoringHandler) LogMetrics(c *gin.Context) {
	var request struct {
		Metrics []struct {
			Name      string  `json:"name"`
			Type      string  `json:"type"`
			StartTime float64 `json:"startTime"`
			Duration  float64 `json:"duration"`
			Timestamp int64   `json:"timestamp"`
		} `json:"metrics" binding:"required"`
		Session struct {
			SessionID string `json:"sessionId"`
			UserID    string `json:"userId,omitempty"`
		} `json:"session,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	for _, metric := range request.Metrics {
		tags := map[string]interface{}{
			"session_id": request.Session.SessionID,
			"user_id":    request.Session.UserID,
			"type":       metric.Type,
		}

		h.logger.LogPerformanceMetric(
			c.Request.Context(),
			metric.Name,
			metric.Duration,
			"ms",
			tags,
		)
	}

	h.logger.Info("Batch metrics logging completed", map[string]interface{}{
		"batch_size": len(request.Metrics),
		"session_id": request.Session.SessionID,
		"user_id":    request.Session.UserID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics logged successfully",
		"count":   len(request.Metrics),
	})
}

// FinalizeSession handles session finalization
func (h *MonitoringHandler) FinalizeSession(c *gin.Context) {
	var request struct {
		Session struct {
			SessionID          string `json:"sessionId" binding:"required"`
			UserID             string `json:"userId,omitempty"`
			StartTime          int64  `json:"startTime"`
			EndTime            int64  `json:"endTime"`
			PageViews          int    `json:"pageViews"`
			ErrorCount         int    `json:"errorCount"`
			PerformanceMetrics struct {
				AvgResponseTime  float64 `json:"avgResponseTime"`
				SlowestOperation string  `json:"slowestOperation"`
				FastestOperation string  `json:"fastestOperation"`
			} `json:"performanceMetrics"`
			UserActions struct {
				Clicks     int `json:"clicks"`
				Keystrokes int `json:"keystrokes"`
				Scrolls    int `json:"scrolls"`
			} `json:"userActions"`
		} `json:"session" binding:"required"`
		SystemHealth struct {
			Status              string          `json:"status"`
			Uptime              int64           `json:"uptime"`
			ErrorRate           float64         `json:"errorRate"`
			PerformanceScore    int             `json:"performanceScore"`
			FeatureAvailability map[string]bool `json:"featureAvailability"`
			LastUpdated         int64           `json:"lastUpdated"`
		} `json:"systemHealth,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	session := request.Session
	duration := session.EndTime - session.StartTime

	// Log session summary
	h.logger.Info("Session finalized", map[string]interface{}{
		"session_id":        session.SessionID,
		"user_id":           session.UserID,
		"duration_ms":       duration,
		"page_views":        session.PageViews,
		"error_count":       session.ErrorCount,
		"avg_response_time": session.PerformanceMetrics.AvgResponseTime,
		"slowest_operation": session.PerformanceMetrics.SlowestOperation,
		"fastest_operation": session.PerformanceMetrics.FastestOperation,
		"user_clicks":       session.UserActions.Clicks,
		"user_keystrokes":   session.UserActions.Keystrokes,
		"user_scrolls":      session.UserActions.Scrolls,
		"system_status":     request.SystemHealth.Status,
		"error_rate":        request.SystemHealth.ErrorRate,
		"performance_score": request.SystemHealth.PerformanceScore,
	})

	// Log performance metrics for the session
	h.logger.LogPerformanceMetric(
		c.Request.Context(),
		"session_duration",
		float64(duration),
		"ms",
		map[string]interface{}{
			"session_id": session.SessionID,
			"user_id":    session.UserID,
		},
	)

	h.logger.LogPerformanceMetric(
		c.Request.Context(),
		"session_error_rate",
		float64(session.ErrorCount)/float64(duration/1000), // errors per second
		"errors/sec",
		map[string]interface{}{
			"session_id": session.SessionID,
			"user_id":    session.UserID,
		},
	)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session finalized successfully",
		"session_id": session.SessionID,
	})
}

// GetHealthStatus returns the health status of the application
func (h *MonitoringHandler) GetHealthStatus(c *gin.Context) {
	ctx := c.Request.Context()

	// Check database connectivity
	dbHealthy := true
	// Note: DatastoreClient doesn't have a Ping method, so we assume it's healthy
	// In a real implementation, you might want to add a specific health check method

	// Get recent error count
	recentErrors := 0
	if entries, err := h.logger.GetLogEntries(ctx, logging.LogFilters{
		Level:     "ERROR",
		StartTime: time.Now().Add(-time.Hour),
		Limit:     100,
	}); err == nil {
		recentErrors = len(entries)
	}

	// Get unresolved alerts count
	unresolvedAlerts := 0
	if alerts, err := h.logger.GetErrorAlerts(ctx, false); err == nil {
		unresolvedAlerts = len(alerts)
	}

	// Determine overall health
	status := "healthy"
	if !dbHealthy || unresolvedAlerts > 10 {
		status = "unhealthy"
	} else if recentErrors > 50 || unresolvedAlerts > 5 {
		status = "degraded"
	}

	health := gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"checks": gin.H{
			"database": gin.H{
				"status": map[bool]string{true: "healthy", false: "unhealthy"}[dbHealthy],
			},
			"errors": gin.H{
				"recent_count": recentErrors,
				"status":       map[bool]string{recentErrors < 50: "healthy", recentErrors >= 50: "unhealthy"}[recentErrors < 50],
			},
			"alerts": gin.H{
				"unresolved_count": unresolvedAlerts,
				"status":           map[bool]string{unresolvedAlerts < 10: "healthy", unresolvedAlerts >= 10: "unhealthy"}[unresolvedAlerts < 10],
			},
		},
	}

	// Set appropriate HTTP status
	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	} else if status == "degraded" {
		httpStatus = http.StatusPartialContent
	}

	c.JSON(httpStatus, health)
}

// GetMetrics returns application metrics
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse time window
	window := time.Hour // Default to 1 hour
	if windowStr := c.Query("window"); windowStr != "" {
		if parsedWindow, err := time.ParseDuration(windowStr); err == nil {
			window = parsedWindow
		}
	}

	// Get log entries for the time window
	entries, err := h.logger.GetLogEntries(ctx, logging.LogFilters{
		StartTime: time.Now().Add(-window),
		Limit:     1000,
	})
	if err != nil {
		h.logger.Error("Failed to get log entries for metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve metrics",
		})
		return
	}

	// Calculate metrics
	metrics := gin.H{
		"window": gin.H{
			"duration": window.String(),
			"start":    time.Now().Add(-window),
			"end":      time.Now(),
		},
		"logs": gin.H{
			"total": len(entries),
			"by_level": gin.H{
				"DEBUG": 0,
				"INFO":  0,
				"WARN":  0,
				"ERROR": 0,
				"FATAL": 0,
			},
			"by_service":   make(map[string]int),
			"by_component": make(map[string]int),
		},
	}

	// Process entries
	byLevel := metrics["logs"].(gin.H)["by_level"].(gin.H)
	byService := metrics["logs"].(gin.H)["by_service"].(map[string]int)
	byComponent := metrics["logs"].(gin.H)["by_component"].(map[string]int)

	for _, entry := range entries {
		// Count by level
		if count, ok := byLevel[entry.Level].(int); ok {
			byLevel[entry.Level] = count + 1
		}

		// Count by service
		byService[entry.Service]++

		// Count by component
		byComponent[entry.Component]++
	}

	c.JSON(http.StatusOK, metrics)
}

// CleanupLogs removes old log entries
func (h *MonitoringHandler) CleanupLogs(c *gin.Context) {
	err := h.logger.CleanupOldLogs(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to cleanup old logs", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cleanup logs",
		})
		return
	}

	h.logger.Info("Log cleanup completed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Log cleanup completed successfully",
	})
}
