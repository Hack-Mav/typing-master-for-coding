package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// CreateFeedback creates a new feedback submission
func CreateFeedback(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			Type             string                 `json:"type" binding:"required"`
			Category         string                 `json:"category" binding:"required"`
			Title            string                 `json:"title" binding:"required"`
			Description      string                 `json:"description" binding:"required"`
			Priority         string                 `json:"priority"`
			Reproducible     bool                   `json:"reproducible"`
			Steps            []string               `json:"steps"`
			ExpectedBehavior string                 `json:"expected_behavior"`
			ActualBehavior   string                 `json:"actual_behavior"`
			Environment      models.EnvironmentInfo `json:"environment"`
			Attachments      []string               `json:"attachments"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		feedback := &models.Feedback{
			ID:               generateID(),
			UserID:           userID,
			Type:             req.Type,
			Category:         req.Category,
			Title:            req.Title,
			Description:      req.Description,
			Priority:         req.Priority,
			Status:           "new",
			Reproducible:     req.Reproducible,
			Steps:            req.Steps,
			ExpectedBehavior: req.ExpectedBehavior,
			ActualBehavior:   req.ActualBehavior,
			Environment:      req.Environment,
			Attachments:      req.Attachments,
			UserAgent:        c.GetHeader("User-Agent"),
			SessionID:        c.GetString("session_id"),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := database.CreateFeedback(db, feedback); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feedback"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"feedback": feedback,
			"message":  "Feedback submitted successfully",
		})
	}
}

// GetFeedback returns feedback submissions (for admins)
func GetFeedback(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")
		category := c.Query("category")
		type_ := c.Query("type")

		feedbacks, err := database.GetFeedback(db, status, category, type_)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feedback"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"feedbacks": feedbacks,
			"total":     len(feedbacks),
		})
	}
}

// GetFeedbackByID returns a specific feedback submission
func GetFeedbackByID(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		feedbackID := c.Param("id")

		feedback, err := database.GetFeedbackByID(db, feedbackID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feedback not found"})
			return
		}

		c.JSON(http.StatusOK, feedback)
	}
}

// UpdateFeedback updates feedback status and details (for admins)
func UpdateFeedback(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		feedbackID := c.Param("id")

		var req struct {
			Status     string `json:"status"`
			AssignedTo string `json:"assigned_to"`
			Priority   string `json:"priority"`
			Resolution string `json:"resolution"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		feedback, err := database.GetFeedbackByID(db, feedbackID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feedback not found"})
			return
		}

		// Update fields
		if req.Status != "" {
			feedback.Status = req.Status
			if req.Status == "resolved" {
				now := time.Now()
				feedback.ResolutionTime = &now
			}
		}
		if req.AssignedTo != "" {
			feedback.AssignedTo = req.AssignedTo
		}
		if req.Priority != "" {
			feedback.Priority = req.Priority
		}
		if req.Resolution != "" {
			feedback.Resolution = req.Resolution
		}
		feedback.UpdatedAt = time.Now()

		if err := database.UpdateFeedback(db, feedback); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feedback"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"feedback": feedback,
			"message":  "Feedback updated successfully",
		})
	}
}

// GetUserFeedback returns feedback submitted by the current user
func GetUserFeedback(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		feedbacks, err := database.GetUserFeedback(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user feedback"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"feedbacks": feedbacks,
			"total":     len(feedbacks),
		})
	}
}

// GetFeedbackCategories returns available feedback categories
func GetFeedbackCategories(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories := []gin.H{
			{"id": "bug_report", "name": "Bug Report", "description": "Report technical issues and errors"},
			{"id": "feature_request", "name": "Feature Request", "description": "Suggest new features and improvements"},
			{"id": "ui_ux", "name": "UI/UX", "description": "Feedback on user interface and experience"},
			{"id": "performance", "name": "Performance", "description": "Report performance issues"},
			{"id": "content", "name": "Content", "description": "Feedback on lessons and tutorials"},
			{"id": "accessibility", "name": "Accessibility", "description": "Accessibility-related feedback"},
			{"id": "general", "name": "General", "description": "General feedback and suggestions"},
		}

		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
		})
	}
}

// GetFeedbackStats returns feedback statistics (for admins)
func GetFeedbackStats(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := database.GetFeedbackStats(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feedback stats"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// CreateSupportTicket creates a new support ticket
func CreateSupportTicket(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			ChannelID   string   `json:"channel_id" binding:"required"`
			Subject     string   `json:"subject" binding:"required"`
			Description string   `json:"description" binding:"required"`
			Priority    string   `json:"priority"`
			Category    string   `json:"category"`
			Tags        []string `json:"tags"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ticket := &models.SupportTicket{
			ID:          generateID(),
			UserID:      userID,
			ChannelID:   req.ChannelID,
			Subject:     req.Subject,
			Description: req.Description,
			Priority:    req.Priority,
			Status:      "open",
			Category:    req.Category,
			Tags:        req.Tags,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := database.CreateSupportTicket(db, ticket); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create support ticket"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"ticket":  ticket,
			"message": "Support ticket created successfully",
		})
	}
}

// GetSupportTickets returns support tickets for the current user
func GetSupportTickets(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		status := c.Query("status")

		tickets, err := database.GetUserSupportTickets(db, userID, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch support tickets"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tickets": tickets,
			"total":   len(tickets),
		})
	}
}

// GetSupportTicketByID returns a specific support ticket
func GetSupportTicketByID(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("id")
		userID := c.GetString("user_id")

		ticket, err := database.GetSupportTicketByID(db, ticketID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Support ticket not found"})
			return
		}

		// Check if user owns the ticket or is admin
		if ticket.UserID != userID {
			role := c.GetString("role")
			if role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}

		c.JSON(http.StatusOK, ticket)
	}
}

// AddSupportMessage adds a message to a support ticket
func AddSupportMessage(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("id")
		userID := c.GetString("user_id")

		var req struct {
			Content     string   `json:"content" binding:"required"`
			IsInternal  bool     `json:"is_internal"`
			Attachments []string `json:"attachments"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ticket, err := database.GetSupportTicketByID(db, ticketID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Support ticket not found"})
			return
		}

		// Check if user owns the ticket or is admin
		if ticket.UserID != userID {
			role := c.GetString("role")
			if role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}

		message := &models.SupportMessage{
			ID:          generateID(),
			TicketID:    ticketID,
			UserID:      userID,
			Content:     req.Content,
			Type:        "user",
			IsInternal:  req.IsInternal,
			Attachments: req.Attachments,
			CreatedAt:   time.Now(),
		}

		if err := database.AddSupportMessage(db, message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add message"})
			return
		}

		// Update ticket
		ticket.UpdatedAt = time.Now()
		if ticket.Status == "open" {
			ticket.Status = "in_progress"
		}
		database.UpdateSupportTicket(db, ticket)

		c.JSON(http.StatusCreated, gin.H{
			"message": message,
			"ticket":  ticket,
		})
	}
}

// GetSupportChannels returns available support channels
func GetSupportChannels(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		channels, err := database.GetActiveSupportChannels(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch support channels"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"channels": channels,
			"total":    len(channels),
		})
	}
}

// RateSupportTicket rates a support ticket (after resolution)
func RateSupportTicket(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("id")
		userID := c.GetString("user_id")

		var req struct {
			Rating   int    `json:"rating" binding:"required,min=1,max=5"`
			Feedback string `json:"feedback"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ticket, err := database.GetSupportTicketByID(db, ticketID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Support ticket not found"})
			return
		}

		// Check if user owns the ticket
		if ticket.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Check if ticket is resolved
		if ticket.Status != "resolved" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Can only rate resolved tickets"})
			return
		}

		ticket.UserRating = &req.Rating
		ticket.UserFeedback = req.Feedback
		ticket.UpdatedAt = time.Now()

		if err := database.UpdateSupportTicket(db, ticket); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rate ticket"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Ticket rated successfully",
			"rating":  req.Rating,
		})
	}
}
