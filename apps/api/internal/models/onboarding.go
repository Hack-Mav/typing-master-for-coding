package models

import (
	"time"
)

// Tutorial represents a step-by-step tutorial for modes and settings
type Tutorial struct {
	ID            string         `json:"id" datastore:"id"`
	Title         string         `json:"title" datastore:"title"`
	Description   string         `json:"description" datastore:"description"`
	Category      string         `json:"category" datastore:"category"`     // "modes", "settings", "features"
	Difficulty    string         `json:"difficulty" datastore:"difficulty"` // "beginner", "intermediate", "advanced"
	Steps         []TutorialStep `json:"steps" datastore:"steps"`
	EstimatedTime int            `json:"estimated_time" datastore:"estimated_time"` // in minutes
	Prerequisites []string       `json:"prerequisites" datastore:"prerequisites"`
	Tags          []string       `json:"tags" datastore:"tags"`
	IsActive      bool           `json:"is_active" datastore:"is_active"`
	CreatedAt     time.Time      `json:"created_at" datastore:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" datastore:"updated_at"`
}

// TutorialStep represents an individual step in a tutorial
type TutorialStep struct {
	ID          string         `json:"id" datastore:"id"`
	Title       string         `json:"title" datastore:"title"`
	Description string         `json:"description" datastore:"description"`
	ContentType string         `json:"content_type" datastore:"content_type"` // "text", "image", "video", "interactive"
	Content     interface{}    `json:"content" datastore:"content"`
	Position    int            `json:"position" datastore:"position"`
	IsRequired  bool           `json:"is_required" datastore:"is_required"`
	Validation  StepValidation `json:"validation" datastore:"validation"`
}

// StepValidation defines validation rules for tutorial steps
type StepValidation struct {
	Type          string      `json:"type" datastore:"type"` // "click", "input", "completion", "time"
	Target        string      `json:"target" datastore:"target"`
	ExpectedValue interface{} `json:"expected_value" datastore:"expected_value"`
	Timeout       int         `json:"timeout" datastore:"timeout"` // in seconds
}

// UserTutorialProgress tracks a user's progress through tutorials
type UserTutorialProgress struct {
	ID             string     `json:"id" datastore:"id"`
	UserID         string     `json:"user_id" datastore:"user_id"`
	TutorialID     string     `json:"tutorial_id" datastore:"tutorial_id"`
	CurrentStep    int        `json:"current_step" datastore:"current_step"`
	CompletedSteps []string   `json:"completed_steps" datastore:"completed_steps"`
	Status         string     `json:"status" datastore:"status"` // "not_started", "in_progress", "completed", "skipped"
	StartedAt      time.Time  `json:"started_at" datastore:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty" datastore:"completed_at"`
	LastAccessedAt time.Time  `json:"last_accessed_at" datastore:"last_accessed_at"`
	TimeSpent      int        `json:"time_spent" datastore:"time_spent"` // in seconds
	SkipReason     string     `json:"skip_reason,omitempty" datastore:"skip_reason"`
}

// Tooltip represents contextual help tooltips
type Tooltip struct {
	ID             string    `json:"id" datastore:"id"`
	Selector       string    `json:"selector" datastore:"selector"` // CSS selector for element
	Title          string    `json:"title" datastore:"title"`
	Content        string    `json:"content" datastore:"content"`
	Position       string    `json:"position" datastore:"position"`     // "top", "bottom", "left", "right", "center"
	Trigger        string    `json:"trigger" datastore:"trigger"`       // "hover", "click", "focus", "auto"
	ShowDelay      int       `json:"show_delay" datastore:"show_delay"` // in milliseconds
	HideDelay      int       `json:"hide_delay" datastore:"hide_delay"` // in milliseconds
	IsPersistent   bool      `json:"is_persistent" datastore:"is_persistent"`
	TargetAudience []string  `json:"target_audience" datastore:"target_audience"` // ["new_users", "all_users"]
	PageContext    string    `json:"page_context" datastore:"page_context"`       // page or component context
	IsActive       bool      `json:"is_active" datastore:"is_active"`
	CreatedAt      time.Time `json:"created_at" datastore:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" datastore:"updated_at"`
}

// HelpArticle represents a help center article
type HelpArticle struct {
	ID              string    `json:"id" datastore:"id"`
	Title           string    `json:"title" datastore:"title"`
	Content         string    `json:"content" datastore:"content"`
	Summary         string    `json:"summary" datastore:"summary"`
	Category        string    `json:"category" datastore:"category"`
	Tags            []string  `json:"tags" datastore:"tags"`
	ViewCount       int       `json:"view_count" datastore:"view_count"`
	HelpfulCount    int       `json:"helpful_count" datastore:"helpful_count"`
	NotHelpfulCount int       `json:"not_helpful_count" datastore:"not_helpful_count"`
	SearchRank      float64   `json:"search_rank" datastore:"search_rank"`
	RelatedArticles []string  `json:"related_articles" datastore:"related_articles"`
	IsActive        bool      `json:"is_active" datastore:"is_active"`
	CreatedAt       time.Time `json:"created_at" datastore:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" datastore:"updated_at"`
}

// FAQ represents a frequently asked question
type FAQ struct {
	ID           string    `json:"id" datastore:"id"`
	Question     string    `json:"question" datastore:"question"`
	Answer       string    `json:"answer" datastore:"answer"`
	Category     string    `json:"category" datastore:"category"`
	Order        int       `json:"order" datastore:"order"`
	ViewCount    int       `json:"view_count" datastore:"view_count"`
	HelpfulCount int       `json:"helpful_count" datastore:"helpful_count"`
	IsActive     bool      `json:"is_active" datastore:"is_active"`
	CreatedAt    time.Time `json:"created_at" datastore:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" datastore:"updated_at"`
}

// Feedback represents user feedback and bug reports
type Feedback struct {
	ID               string          `json:"id" datastore:"id"`
	UserID           string          `json:"user_id" datastore:"user_id"`
	Type             string          `json:"type" datastore:"type"` // "bug_report", "feature_request", "general_feedback"
	Category         string          `json:"category" datastore:"category"`
	Title            string          `json:"title" datastore:"title"`
	Description      string          `json:"description" datastore:"description"`
	Priority         string          `json:"priority" datastore:"priority"` // "low", "medium", "high", "critical"
	Status           string          `json:"status" datastore:"status"`     // "new", "in_progress", "resolved", "closed"
	Reproducible     bool            `json:"reproducible" datastore:"reproducible"`
	Steps            []string        `json:"steps" datastore:"steps"`
	ExpectedBehavior string          `json:"expected_behavior" datastore:"expected_behavior"`
	ActualBehavior   string          `json:"actual_behavior" datastore:"actual_behavior"`
	Environment      EnvironmentInfo `json:"environment" datastore:"environment"`
	Attachments      []string        `json:"attachments" datastore:"attachments"`
	UserAgent        string          `json:"user_agent" datastore:"user_agent"`
	SessionID        string          `json:"session_id" datastore:"session_id"`
	AssignedTo       string          `json:"assigned_to,omitempty" datastore:"assigned_to"`
	Resolution       string          `json:"resolution,omitempty" datastore:"resolution"`
	ResolutionTime   *time.Time      `json:"resolution_time,omitempty" datastore:"resolution_time"`
	CreatedAt        time.Time       `json:"created_at" datastore:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" datastore:"updated_at"`
}

// EnvironmentInfo captures environment details for bug reports
type EnvironmentInfo struct {
	Browser        string `json:"browser" datastore:"browser"`
	BrowserVersion string `json:"browser_version" datastore:"browser_version"`
	OS             string `json:"os" datastore:"os"`
	DeviceType     string `json:"device_type" datastore:"device_type"`
	ScreenSize     string `json:"screen_size" datastore:"screen_size"`
	Language       string `json:"language" datastore:"language"`
	Timezone       string `json:"timezone" datastore:"timezone"`
}

// UserOnboardingState tracks overall onboarding progress
type UserOnboardingState struct {
	ID                 string                `json:"id" datastore:"id"`
	UserID             string                `json:"user_id" datastore:"user_id"`
	CompletedTutorials []string              `json:"completed_tutorials" datastore:"completed_tutorials"`
	ViewedTooltips     []string              `json:"viewed_tooltips" datastore:"viewed_tooltips"`
	ReadArticles       []string              `json:"read_articles" datastore:"read_articles"`
	OverallProgress    float64               `json:"overall_progress" datastore:"overall_progress"` // 0.0 to 1.0
	LastActivityAt     time.Time             `json:"last_activity_at" datastore:"last_activity_at"`
	OnboardingVersion  string                `json:"onboarding_version" datastore:"onboarding_version"`
	SkippedSteps       []string              `json:"skipped_steps" datastore:"skipped_steps"`
	Preferences        OnboardingPreferences `json:"preferences" datastore:"preferences"`
	CreatedAt          time.Time             `json:"created_at" datastore:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at" datastore:"updated_at"`
}

// OnboardingPreferences represents user preferences for onboarding
type OnboardingPreferences struct {
	ShowTooltips       bool   `json:"show_tooltips" datastore:"show_tooltips"`
	AutoStartTutorials bool   `json:"auto_start_tutorials" datastore:"auto_start_tutorials"`
	InteractiveMode    bool   `json:"interactive_mode" datastore:"interactive_mode"`
	CompactView        bool   `json:"compact_view" datastore:"compact_view"`
	Language           string `json:"language" datastore:"language"`
}
