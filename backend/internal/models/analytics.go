package models

import (
	"time"
)

// AnalyticsEvent represents a single analytics event
type AnalyticsEvent struct {
	ID         string                 `json:"id" datastore:"id"`
	UserID     string                 `json:"user_id,omitempty" datastore:"user_id"`
	SessionID  string                 `json:"session_id" datastore:"session_id"`
	EventType  string                 `json:"event_type" datastore:"event_type"`
	EventName  string                 `json:"event_name" datastore:"event_name"`
	Properties map[string]interface{} `json:"properties" datastore:"properties"`
	Timestamp  time.Time              `json:"timestamp" datastore:"timestamp"`
	UserAgent  string                 `json:"user_agent" datastore:"user_agent"`
	IPAddress  string                 `json:"ip_address" datastore:"ip_address"`
	Country    string                 `json:"country" datastore:"country"`
	DeviceType string                 `json:"device_type" datastore:"device_type"`
	Browser    string                 `json:"browser" datastore:"browser"`
	Version    string                 `json:"version" datastore:"version"`
}

// AnalyticsConsent represents user consent for analytics
type AnalyticsConsent struct {
	ID             string    `json:"id" datastore:"id"`
	UserID         string    `json:"user_id" datastore:"user_id"`
	ConsentGiven   bool      `json:"consent_given" datastore:"consent_given"`
	ConsentType    string    `json:"consent_type" datastore:"consent_type"` // "full", "minimal", "none"
	ConsentVersion string    `json:"consent_version" datastore:"consent_version"`
	DataRetention  int       `json:"data_retention" datastore:"data_retention"` // days
	AllowedEvents  []string  `json:"allowed_events" datastore:"allowed_events"`
	DeclinedEvents []string  `json:"declined_events" datastore:"declined_events"`
	IPAddressHash  string    `json:"ip_address_hash" datastore:"ip_address_hash"`
	UserAgentHash  string    `json:"user_agent_hash" datastore:"user_agent_hash"`
	CreatedAt      time.Time `json:"created_at" datastore:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" datastore:"updated_at"`
	ExpiresAt      time.Time `json:"expires_at" datastore:"expires_at"`
}

// ABTest represents an A/B test configuration for user support features
type UserSupportABTest struct {
	ID                         string         `json:"id" datastore:"id"`
	Name                       string         `json:"name" datastore:"name"`
	Description                string         `json:"description" datastore:"description"`
	Status                     string         `json:"status" datastore:"status"` // "draft", "running", "paused", "completed"
	TrafficSplit               []TrafficSplit `json:"traffic_split" datastore:"traffic_split"`
	TargetAudience             TargetAudience `json:"target_audience" datastore:"target_audience"`
	SuccessMetrics             []string       `json:"success_metrics" datastore:"success_metrics"`
	StartDate                  time.Time      `json:"start_date" datastore:"start_date"`
	EndDate                    *time.Time     `json:"end_date,omitempty" datastore:"end_date"`
	SampleSize                 int            `json:"sample_size" datastore:"sample_size"`
	ConfidenceLevel            float64        `json:"confidence_level" datastore:"confidence_level"`
	IsStatisticallySignificant bool           `json:"is_statistically_significant" datastore:"is_statistically_significant"`
	Winner                     *string        `json:"winner,omitempty" datastore:"winner"`
	CreatedAt                  time.Time      `json:"created_at" datastore:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at" datastore:"updated_at"`
}

// TrafficSplit represents traffic distribution for A/B test variants
type TrafficSplit struct {
	VariantID     string                 `json:"variant_id" datastore:"variant_id"`
	VariantName   string                 `json:"variant_name" datastore:"variant_name"`
	Weight        float64                `json:"weight" datastore:"weight"` // 0.0 to 1.0
	Configuration map[string]interface{} `json:"configuration" datastore:"configuration"`
	IsControl     bool                   `json:"is_control" datastore:"is_control"`
}

// TargetAudience defines who should be included in the test
type TargetAudience struct {
	Filters    []AudienceFilter `json:"filters" datastore:"filters"`
	Percentage float64          `json:"percentage" datastore:"percentage"` // 0.0 to 1.0
}

// AudienceFilter represents a filter for test audience
type AudienceFilter struct {
	Field    string      `json:"field" datastore:"field"`
	Operator string      `json:"operator" datastore:"operator"` // "equals", "contains", "greater_than", "less_than", "in"
	Value    interface{} `json:"value" datastore:"value"`
}

// UserABTestAssignment tracks which variant a user is assigned to
type UserABTestAssignment struct {
	ID              string    `json:"id" datastore:"id"`
	UserID          string    `json:"user_id" datastore:"user_id"`
	ABTestID        string    `json:"ab_test_id" datastore:"ab_test_id"`
	VariantID       string    `json:"variant_id" datastore:"variant_id"`
	AssignedAt      time.Time `json:"assigned_at" datastore:"assigned_at"`
	IsExcluded      bool      `json:"is_excluded" datastore:"is_excluded"`
	ExclusionReason string    `json:"exclusion_reason,omitempty" datastore:"exclusion_reason"`
}

// ABTestResult represents aggregated results for an A/B test
type ABTestResult struct {
	ID                      string    `json:"id" datastore:"id"`
	ABTestID                string    `json:"ab_test_id" datastore:"ab_test_id"`
	VariantID               string    `json:"variant_id" datastore:"variant_id"`
	MetricName              string    `json:"metric_name" datastore:"metric_name"`
	MetricValue             float64   `json:"metric_value" datastore:"metric_value"`
	SampleSize              int       `json:"sample_size" datastore:"sample_size"`
	StandardError           float64   `json:"standard_error" datastore:"standard_error"`
	ConfidenceIntervalLow   float64   `json:"confidence_interval_low" datastore:"confidence_interval_low"`
	ConfidenceIntervalHigh  float64   `json:"confidence_interval_high" datastore:"confidence_interval_high"`
	PValue                  float64   `json:"p_value" datastore:"p_value"`
	StatisticalSignificance bool      `json:"statistical_significance" datastore:"statistical_significance"`
	PeriodStart             time.Time `json:"period_start" datastore:"period_start"`
	PeriodEnd               time.Time `json:"period_end" datastore:"period_end"`
	CreatedAt               time.Time `json:"created_at" datastore:"created_at"`
}

// CommunityForum represents a community forum or discussion board
type CommunityForum struct {
	ID           string    `json:"id" datastore:"id"`
	Name         string    `json:"name" datastore:"name"`
	Description  string    `json:"description" datastore:"description"`
	Category     string    `json:"category" datastore:"category"`
	Type         string    `json:"type" datastore:"type"` // "general", "technical", "feature_requests", "bug_reports", "announcements"
	IsModerated  bool      `json:"is_moderated" datastore:"is_moderated"`
	PostCount    int       `json:"post_count" datastore:"post_count"`
	MemberCount  int       `json:"member_count" datastore:"member_count"`
	LastActivity time.Time `json:"last_activity" datastore:"last_activity"`
	IsActive     bool      `json:"is_active" datastore:"is_active"`
	CreatedAt    time.Time `json:"created_at" datastore:"updated_at" datastore:"updated_at"`
}

// ForumPost represents a post in a community forum
type ForumPost struct {
	ID          string     `json:"id" datastore:"id"`
	ForumID     string     `json:"forum_id" datastore:"forum_id"`
	UserID      string     `json:"user_id" datastore:"user_id"`
	Title       string     `json:"title" datastore:"title"`
	Content     string     `json:"content" datastore:"content"`
	Type        string     `json:"type" datastore:"type"`     // "question", "discussion", "announcement", "feature_request"
	Status      string     `json:"status" datastore:"status"` // "open", "closed", "locked", "pinned"
	Tags        []string   `json:"tags" datastore:"tags"`
	ViewCount   int        `json:"view_count" datastore:"view_count"`
	LikeCount   int        `json:"like_count" datastore:"like_count"`
	ReplyCount  int        `json:"reply_count" datastore:"reply_count"`
	IsPinned    bool       `json:"is_pinned" datastore:"is_pinned"`
	IsLocked    bool       `json:"is_locked" datastore:"is_locked"`
	LastReplyAt *time.Time `json:"last_reply_at,omitempty" datastore:"last_reply_at"`
	CreatedAt   time.Time  `json:"created_at" datastore:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" datastore:"updated_at"`
}

// ForumReply represents a reply to a forum post
type ForumReply struct {
	ID        string    `json:"id" datastore:"id"`
	PostID    string    `json:"post_id" datastore:"post_id"`
	UserID    string    `json:"user_id" datastore:"user_id"`
	Content   string    `json:"content" datastore:"content"`
	LikeCount int       `json:"like_count" datastore:"like_count"`
	IsAnswer  bool      `json:"is_answer" datastore:"is_answer"` // marked as best answer for questions
	CreatedAt time.Time `json:"created_at" datastore:"created_at"`
	UpdatedAt time.Time `json:"updated_at" datastore:"updated_at"`
}

// SupportChannel represents a support channel (chat, email, etc.)
type SupportChannel struct {
	ID           string                 `json:"id" datastore:"id"`
	Name         string                 `json:"name" datastore:"name"`
	Type         string                 `json:"type" datastore:"type"` // "chat", "email", "phone", "forum", "knowledge_base"
	Description  string                 `json:"description" datastore:"description"`
	Availability string                 `json:"availability" datastore:"availability"`   // "24/7", "business_hours", "limited"
	ResponseTime string                 `json:"response_time" datastore:"response_time"` // "instant", "minutes", "hours", "days"
	IsActive     bool                   `json:"is_active" datastore:"is_active"`
	Config       map[string]interface{} `json:"config" datastore:"config"`
	CreatedAt    time.Time              `json:"created_at" datastore:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" datastore:"updated_at"`
}

// SupportTicket represents a support ticket or conversation
type SupportTicket struct {
	ID                string           `json:"id" datastore:"id"`
	UserID            string           `json:"user_id" datastore:"user_id"`
	ChannelID         string           `json:"channel_id" datastore:"channel_id"`
	Subject           string           `json:"subject" datastore:"subject"`
	Description       string           `json:"description" datastore:"description"`
	Priority          string           `json:"priority" datastore:"priority"` // "low", "medium", "high", "urgent"
	Status            string           `json:"status" datastore:"status"`     // "open", "in_progress", "waiting", "resolved", "closed"
	Category          string           `json:"category" datastore:"category"`
	AssignedTo        string           `json:"assigned_to,omitempty" datastore:"assigned_to"`
	Tags              []string         `json:"tags" datastore:"tags"`
	Messages          []SupportMessage `json:"messages" datastore:"messages"`
	UserRating        *int             `json:"user_rating,omitempty" datastore:"user_rating"` // 1-5 stars
	UserFeedback      string           `json:"user_feedback,omitempty" datastore:"user_feedback"`
	ResolutionTime    *time.Time       `json:"resolution_time,omitempty" datastore:"resolution_time"`
	FirstResponseTime *time.Time       `json:"first_response_time,omitempty" datastore:"first_response_time"`
	CreatedAt         time.Time        `json:"created_at" datastore:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" datastore:"updated_at"`
	ClosedAt          *time.Time       `json:"closed_at,omitempty" datastore:"closed_at"`
}

// SupportMessage represents a message in a support conversation
type SupportMessage struct {
	ID          string    `json:"id" datastore:"id"`
	TicketID    string    `json:"ticket_id" datastore:"ticket_id"`
	UserID      string    `json:"user_id" datastore:"user_id"`
	Content     string    `json:"content" datastore:"content"`
	Type        string    `json:"type" datastore:"type"`               // "user", "agent", "system"
	IsInternal  bool      `json:"is_internal" datastore:"is_internal"` // internal notes
	Attachments []string  `json:"attachments" datastore:"attachments"`
	CreatedAt   time.Time `json:"created_at" datastore:"created_at"`
}
