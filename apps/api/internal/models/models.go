package models

import (
	"time"
)

// User entity for Datastore
type User struct {
	ID                    string                 `datastore:"-" json:"id"`
	Handle                string                 `datastore:"handle" json:"handle"`
	Email                 string                 `datastore:"email" json:"email"`
	PasswordHash          string                 `datastore:"password_hash" json:"-"` // Never expose in JSON
	Role                  string                 `datastore:"role" json:"role"`       // "user", "admin", "moderator"
	IsAnonymous           bool                   `datastore:"is_anonymous" json:"is_anonymous"`
	Locale                string                 `datastore:"locale" json:"locale"`
	KeyboardLayout        string                 `datastore:"keyboard_layout" json:"keyboard_layout"`
	PrivacyMode           bool                   `datastore:"privacy_mode" json:"privacy_mode"`
	TelemetryConsent      bool                   `datastore:"telemetry_consent" json:"telemetry_consent"`
	DataProcessingConsent bool                   `datastore:"data_processing_consent" json:"data_processing_consent"`
	Settings              map[string]interface{} `datastore:"settings" json:"settings"`
	CreatedAt             time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt             time.Time              `datastore:"updated_at" json:"updated_at"`
	LastLoginAt           *time.Time             `datastore:"last_login_at" json:"last_login_at,omitempty"`

	// MFA fields
	MFAEnabled     bool       `datastore:"mfa_enabled" json:"mfa_enabled"`
	MFASecret      string     `datastore:"mfa_secret" json:"-"`       // Never expose in JSON
	MFABackupCodes []string   `datastore:"mfa_backup_codes" json:"-"` // Never expose in JSON
	MFASetupAt     *time.Time `datastore:"mfa_setup_at" json:"mfa_setup_at,omitempty"`
}

// Language entity for Datastore
type Language struct {
	ID              string                 `datastore:"-" json:"id"`
	Name            string                 `datastore:"name" json:"name"`
	Version         int                    `datastore:"version" json:"version"`
	ParserID        string                 `datastore:"parser_id" json:"parser_id"`
	GrammarConfig   map[string]interface{} `datastore:"grammar_config" json:"grammar_config"`
	WhitespaceRules map[string]interface{} `datastore:"whitespace_rules" json:"whitespace_rules"`
	CreatedBy       string                 `datastore:"created_by" json:"created_by"`
	CreatedAt       time.Time              `datastore:"created_at" json:"created_at"`
}

// Lesson entity for Datastore
type Lesson struct {
	ID               string    `datastore:"-" json:"id"`
	LanguageID       string    `datastore:"language_id" json:"language_id"`
	Title            string    `datastore:"title" json:"title"`
	Difficulty       int       `datastore:"difficulty" json:"difficulty"`
	Objectives       []string  `datastore:"objectives" json:"objectives"`
	Prerequisites    []string  `datastore:"prerequisites" json:"prerequisites"`
	EstimatedMinutes int       `datastore:"estimated_minutes" json:"estimated_minutes"`
	TokensCovered    []string  `datastore:"tokens_covered" json:"tokens_covered"`
	SnippetIDs       []string  `datastore:"snippet_ids" json:"snippet_ids"`
	Version          int       `datastore:"version" json:"version"`
	CreatedBy        string    `datastore:"created_by" json:"created_by"`
	CreatedAt        time.Time `datastore:"created_at" json:"created_at"`
}

// Snippet entity for Datastore
type Snippet struct {
	ID                string                 `datastore:"-" json:"id"`
	LanguageID        string                 `datastore:"language_id" json:"language_id"`
	Title             string                 `datastore:"title" json:"title"`
	SourceCode        string                 `datastore:"source_code,noindex" json:"source_code"`
	Tags              []string               `datastore:"tags" json:"tags"`
	Difficulty        int                    `datastore:"difficulty" json:"difficulty"`
	EstimatedTime     int                    `datastore:"estimated_time" json:"estimated_time"`
	Checksum          string                 `datastore:"checksum" json:"checksum"`
	AccessibilityTags map[string]interface{} `datastore:"accessibility_tags" json:"accessibility_tags"`
	CreatedBy         string                 `datastore:"created_by" json:"created_by"`
	Version           int                    `datastore:"version" json:"version"`
	CreatedAt         time.Time              `datastore:"created_at" json:"created_at"`
}

// Session entity for Datastore
type Session struct {
	ID         string                 `datastore:"-" json:"id"`
	UserID     string                 `datastore:"user_id" json:"user_id"`
	Mode       string                 `datastore:"mode" json:"mode"`
	LanguageID string                 `datastore:"language_id" json:"language_id"`
	LessonID   string                 `datastore:"lesson_id" json:"lesson_id"`
	SnippetID  string                 `datastore:"snippet_id" json:"snippet_id"`
	StartedAt  time.Time              `datastore:"started_at" json:"started_at"`
	EndedAt    *time.Time             `datastore:"ended_at" json:"ended_at"`
	DurationMs int64                  `datastore:"duration_ms" json:"duration_ms"`
	Settings   map[string]interface{} `datastore:"settings" json:"settings"`
	CreatedAt  time.Time              `datastore:"created_at" json:"created_at"`
}

// SessionEvent entity for Datastore
type SessionEvent struct {
	ID             string                 `datastore:"-" json:"id"`
	SessionID      string                 `datastore:"session_id" json:"session_id"`
	TimestampMs    int64                  `datastore:"timestamp_ms" json:"timestamp_ms"`
	KeyPressed     string                 `datastore:"key_pressed" json:"key_pressed"`
	Action         string                 `datastore:"action" json:"action"`
	CursorPosition int                    `datastore:"cursor_position" json:"cursor_position"`
	ErrorFlag      bool                   `datastore:"error_flag" json:"error_flag"`
	ExpectedToken  string                 `datastore:"expected_token" json:"expected_token"`
	Metadata       map[string]interface{} `datastore:"metadata,noindex" json:"metadata"`
	CreatedAt      time.Time              `datastore:"created_at" json:"created_at"`
}

// Result entity for Datastore
type Result struct {
	SessionID      string                 `datastore:"-" json:"session_id"`
	CPM            float64                `datastore:"cpm" json:"cpm"`
	TWPM           float64                `datastore:"twpm" json:"twpm"`
	RawAccuracy    float64                `datastore:"raw_accuracy" json:"raw_accuracy"`
	TokenAccuracy  float64                `datastore:"token_accuracy" json:"token_accuracy"`
	SyntaxAccuracy float64                `datastore:"syntax_accuracy" json:"syntax_accuracy"`
	BackspaceRate  float64                `datastore:"backspace_rate" json:"backspace_rate"`
	CompositeScore int                    `datastore:"composite_score" json:"composite_score"`
	Breakdown      map[string]interface{} `datastore:"breakdown" json:"breakdown"`
	CreatedAt      time.Time              `datastore:"created_at" json:"created_at"`
}

// Playlist entity for Datastore
// Playlists are collections of lessons or snippets for custom practice
type Playlist struct {
	ID          string    `datastore:"-" json:"id"`
	UserID      string    `datastore:"user_id" json:"user_id"`
	Title       string    `datastore:"title" json:"title"`
	Description string    `datastore:"description" json:"description"`
	LanguageID  string    `datastore:"language_id" json:"language_id"`
	LessonIDs   []string  `datastore:"lesson_ids" json:"lesson_ids"`
	SnippetIDs  []string  `datastore:"snippet_ids" json:"snippet_ids"`
	Tags        []string  `datastore:"tags" json:"tags"`
	IsPublic    bool      `datastore:"is_public" json:"is_public"`
	Difficulty  int       `datastore:"difficulty" json:"difficulty"`
	Version     int       `datastore:"version" json:"version"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

// ContentVersion entity for Datastore
// Tracks versions of content for audit and rollback capabilities
type ContentVersion struct {
	ID          string                 `datastore:"-" json:"id"`
	ContentType string                 `datastore:"content_type" json:"content_type"` // "language", "lesson", "snippet", "playlist"
	ContentID   string                 `datastore:"content_id" json:"content_id"`
	Version     int                    `datastore:"version" json:"version"`
	Content     map[string]interface{} `datastore:"content" json:"content"`
	Checksum    string                 `datastore:"checksum" json:"checksum"`
	CreatedBy   string                 `datastore:"created_by" json:"created_by"`
	ChangeNotes string                 `datastore:"change_notes" json:"change_notes"`
	IsActive    bool                   `datastore:"is_active" json:"is_active"`
	CreatedAt   time.Time              `datastore:"created_at" json:"created_at"`
}

// ContentValidation entity for Datastore
// Stores validation results and checksums for content integrity
type ContentValidation struct {
	ID               string    `datastore:"-" json:"id"`
	ContentType      string    `datastore:"content_type" json:"content_type"` // "language", "lesson", "snippet", "playlist"
	ContentID        string    `datastore:"content_id" json:"content_id"`
	Checksum         string    `datastore:"checksum" json:"checksum"`
	ValidationStatus string    `datastore:"validation_status" json:"validation_status"` // "valid", "invalid", "warning"
	ValidationErrors []string  `datastore:"validation_errors" json:"validation_errors"`
	ValidatedAt      time.Time `datastore:"validated_at" json:"validated_at"`
}

// LessonProgress entity for Datastore
// Tracks user progress through lessons and lesson flows
type LessonProgress struct {
	ID              string    `datastore:"-" json:"id"`
	UserID          string    `datastore:"user_id" json:"user_id"`
	LessonID        string    `datastore:"lesson_id" json:"lesson_id"`
	CurrentStage    string    `datastore:"current_stage" json:"current_stage"` // "intro", "core", "idioms", "advanced", "review"
	CompletedStages []string  `datastore:"completed_stages" json:"completed_stages"`
	ProgressPercent float64   `datastore:"progress_percent" json:"progress_percent"`
	TokensCovered   []string  `datastore:"tokens_covered" json:"tokens_covered"`
	IsUnlocked      bool      `datastore:"is_unlocked" json:"is_unlocked"`
	IsCompleted     bool      `datastore:"is_completed" json:"is_completed"`
	BestScore       int       `datastore:"best_score" json:"best_score"`
	AttemptsCount   int       `datastore:"attempts_count" json:"attempts_count"`
	LastAttemptAt   time.Time `datastore:"last_attempt_at" json:"last_attempt_at"`
	CreatedAt       time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt       time.Time `datastore:"updated_at" json:"updated_at"`
}

// DifficultyAssessment entity for Datastore
// Stores difficulty assessments and adaptive progression data
type DifficultyAssessment struct {
	ID                 string                 `datastore:"-" json:"id"`
	ContentType        string                 `datastore:"content_type" json:"content_type"` // "lesson", "snippet"
	ContentID          string                 `datastore:"content_id" json:"content_id"`
	AssessedDifficulty int                    `datastore:"assessed_difficulty" json:"assessed_difficulty"`
	ComplexityScore    float64                `datastore:"complexity_score" json:"complexity_score"`
	TokenComplexity    map[string]float64     `datastore:"token_complexity" json:"token_complexity"`
	EstimatedTime      int                    `datastore:"estimated_time" json:"estimated_time"` // minutes
	SuccessRate        float64                `datastore:"success_rate" json:"success_rate"`
	AverageAttempts    float64                `datastore:"average_attempts" json:"average_attempts"`
	AdaptiveLevel      int                    `datastore:"adaptive_level" json:"adaptive_level"`
	AssessmentMetadata map[string]interface{} `datastore:"assessment_metadata" json:"assessment_metadata"`
	AssessedAt         time.Time              `datastore:"assessed_at" json:"assessed_at"`
}

// ScoringMetrics entity for Datastore (extends Result with more detailed metrics)
type ScoringMetrics struct {
	ID         string `datastore:"-" json:"id"`
	SessionID  string `datastore:"session_id" json:"session_id"`
	UserID     string `datastore:"user_id" json:"user_id"`
	LanguageID string `datastore:"language_id" json:"language_id"`
	Mode       string `datastore:"mode" json:"mode"`

	// Speed metrics
	CPM  float64 `datastore:"cpm" json:"cpm"`
	TWPM float64 `datastore:"twpm" json:"twpm"`
	KPS  float64 `datastore:"kps" json:"kps"`

	// Accuracy metrics
	RawAccuracy        float64 `datastore:"raw_accuracy" json:"raw_accuracy"`
	TokenAccuracy      float64 `datastore:"token_accuracy" json:"token_accuracy"`
	SyntaxAccuracy     float64 `datastore:"syntax_accuracy" json:"syntax_accuracy"`
	WhitespaceAccuracy float64 `datastore:"whitespace_accuracy" json:"whitespace_accuracy"`

	// Efficiency metrics
	BackspaceRate   float64 `datastore:"backspace_rate" json:"backspace_rate"`
	CorrectionRate  float64 `datastore:"correction_rate" json:"correction_rate"`
	IdleTimePercent float64 `datastore:"idle_time_percent" json:"idle_time_percent"`

	// Composite scores
	CompositeScore   int     `datastore:"composite_score" json:"composite_score"`
	ConsistencyScore float64 `datastore:"consistency_score" json:"consistency_score"`
	EfficiencyScore  float64 `datastore:"efficiency_score" json:"efficiency_score"`

	// Detailed breakdown
	ErrorClusters       map[string]interface{} `datastore:"error_clusters" json:"error_clusters"`
	PerformanceInsights map[string]interface{} `datastore:"performance_insights" json:"performance_insights"`
	TypingPatterns      map[string]interface{} `datastore:"typing_patterns" json:"typing_patterns"`

	// Anti-cheat flags
	SuspiciousActivity bool     `datastore:"suspicious_activity" json:"suspicious_activity"`
	CheatFlags         []string `datastore:"cheat_flags" json:"cheat_flags"`
	ConfidenceScore    float64  `datastore:"confidence_score" json:"confidence_score"`

	// Timestamps
	CalculatedAt time.Time `datastore:"calculated_at" json:"calculated_at"`
	CreatedAt    time.Time `datastore:"created_at" json:"created_at"`
}

// Leaderboard entity for Datastore
type Leaderboard struct {
	ID         string `datastore:"-" json:"id"`
	UserID     string `datastore:"user_id" json:"user_id"`
	SessionID  string `datastore:"session_id" json:"session_id"`
	LanguageID string `datastore:"language_id" json:"language_id"`
	Mode       string `datastore:"mode" json:"mode"`
	Scope      string `datastore:"scope" json:"scope"`             // "global", "friends", "organization"
	TimeWindow string `datastore:"time_window" json:"time_window"` // "daily", "weekly", "monthly", "all_time"

	// Rankings
	Rank     int     `datastore:"rank" json:"rank"`
	Score    int     `datastore:"score" json:"score"`
	CPM      float64 `datastore:"cpm" json:"cpm"`
	TWPM     float64 `datastore:"twpm" json:"twpm"`
	Accuracy float64 `datastore:"accuracy" json:"accuracy"`

	// Metadata
	MetricsSnapshot    map[string]interface{} `datastore:"metrics_snapshot" json:"metrics_snapshot"`
	Badge              string                 `datastore:"badge" json:"badge"`
	IsVerified         bool                   `datastore:"is_verified" json:"is_verified"`
	VerificationMethod string                 `datastore:"verification_method" json:"verification_method"`

	// Timestamps
	RecordedAt  time.Time `datastore:"recorded_at" json:"recorded_at"`
	WindowStart time.Time `datastore:"window_start" json:"window_start"`
	WindowEnd   time.Time `datastore:"window_end" json:"window_end"`
}

// AntiCheatReport entity for Datastore
type AntiCheatReport struct {
	ID        string `datastore:"-" json:"id"`
	SessionID string `datastore:"session_id" json:"session_id"`
	UserID    string `datastore:"user_id" json:"user_id"`

	// Detection results
	PasteDetected   bool `datastore:"paste_detected" json:"paste_detected"`
	UnrealisticKPS  bool `datastore:"unrealistic_kps" json:"unrealistic_kps"`
	AutoTypePattern bool `datastore:"auto_type_pattern" json:"auto_type_pattern"`
	WindowFocusLost bool `datastore:"window_focus_lost" json:"window_focus_lost"`
	AnomalousTiming bool `datastore:"anomalous_timing" json:"anomalous_timing"`

	// Analysis data
	KPSVariance        float64 `datastore:"kps_variance" json:"kps_variance"`
	BurstConsistency   float64 `datastore:"burst_consistency" json:"burst_consistency"`
	ErrorPatternScore  float64 `datastore:"error_pattern_score" json:"error_pattern_score"`
	StatisticalAnomaly float64 `datastore:"statistical_anomaly" json:"statistical_anomaly"`

	// Evidence
	SuspiciousEvents []string               `datastore:"suspicious_events" json:"suspicious_events"`
	EvidenceDetails  map[string]interface{} `datastore:"evidence_details" json:"evidence_details"`
	ConfidenceScore  float64                `datastore:"confidence_score" json:"confidence_score"`
	RiskLevel        string                 `datastore:"risk_level" json:"risk_level"` // "low", "medium", "high", "critical"

	// Actions taken
	FlaggedForReview    bool `datastore:"flagged_for_review" json:"flagged_for_review"`
	ScoreInvalidated    bool `datastore:"score_invalidated" json:"score_invalidated"`
	LeaderboardExcluded bool `datastore:"leaderboard_excluded" json:"leaderboard_excluded"`

	// Timestamps
	DetectedAt time.Time  `datastore:"detected_at" json:"detected_at"`
	ReviewedAt *time.Time `datastore:"reviewed_at" json:"reviewed_at"`
	ReviewedBy string     `datastore:"reviewed_by" json:"reviewed_by"`
}

// Tournament entity for Datastore
type Tournament struct {
	ID          string `datastore:"-" json:"id"`
	Title       string `datastore:"title" json:"title"`
	Description string `datastore:"description" json:"description"`
	LanguageID  string `datastore:"language_id" json:"language_id"`
	Mode        string `datastore:"mode" json:"mode"`

	// Tournament settings
	VerificationRequired bool   `datastore:"verification_required" json:"verification_required"`
	VerificationMethod   string `datastore:"verification_method" json:"verification_method"` // "webcam", "hid", "both"
	AntiCheatLevel       string `datastore:"anti_cheat_level" json:"anti_cheat_level"`       // "standard", "enhanced", "maximum"

	// Schedule
	RegistrationStart time.Time `datastore:"registration_start" json:"registration_start"`
	RegistrationEnd   time.Time `datastore:"registration_end" json:"registration_end"`
	StartTime         time.Time `datastore:"start_time" json:"start_time"`
	EndTime           time.Time `datastore:"end_time" json:"end_time"`
	DurationMinutes   int       `datastore:"duration_minutes" json:"duration_minutes"`

	// Participation
	MaxParticipants     int      `datastore:"max_participants" json:"max_participants"`
	CurrentParticipants int      `datastore:"current_participants" json:"current_participants"`
	ParticipantIDs      []string `datastore:"participant_ids" json:"participant_ids"`

	// Prizes and rewards
	PrizePool         float64            `datastore:"prize_pool" json:"prize_pool"`
	PrizeDistribution map[string]float64 `datastore:"prize_distribution" json:"prize_distribution"`
	BadgeReward       string             `datastore:"badge_reward" json:"badge_reward"`

	// Status
	Status string `datastore:"status" json:"status"` // "upcoming", "registration", "active", "completed", "cancelled"

	// Results
	FinalRankings []string `datastore:"final_rankings" json:"final_rankings"`
	Winners       []string `datastore:"winners" json:"winners"`

	// Metadata
	CreatedBy string    `datastore:"created_by" json:"created_by"`
	CreatedAt time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt time.Time `datastore:"updated_at" json:"updated_at"`
}

// TournamentParticipant entity for Datastore
type TournamentParticipant struct {
	ID           string `datastore:"-" json:"id"`
	TournamentID string `datastore:"tournament_id" json:"tournament_id"`
	UserID       string `datastore:"user_id" json:"user_id"`

	// Registration info
	RegisteredAt       time.Time              `datastore:"registered_at" json:"registered_at"`
	VerificationStatus string                 `datastore:"verification_status" json:"verification_status"` // "pending", "verified", "failed"
	VerificationData   map[string]interface{} `datastore:"verification_data" json:"verification_data"`

	// Tournament results
	SessionID   string     `datastore:"session_id" json:"session_id"`
	Score       int        `datastore:"score" json:"score"`
	Rank        int        `datastore:"rank" json:"rank"`
	CompletedAt *time.Time `datastore:"completed_at" json:"completed_at"`

	// Status
	Status                 string `datastore:"status" json:"status"` // "registered", "active", "completed", "disqualified"
	DisqualificationReason string `datastore:"disqualification_reason" json:"disqualification_reason"`

	CreatedAt time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt time.Time `datastore:"updated_at" json:"updated_at"`
}

// ScoringConfig entity for Datastore
// Stores configurable scoring weights and parameters
type ScoringConfig struct {
	ID      string `datastore:"-" json:"id"`
	Version string `datastore:"version" json:"version"`

	// Weight configurations
	TWPMWeight           float64 `datastore:"twpm_weight" json:"twpm_weight"`
	RawAccuracyWeight    float64 `datastore:"raw_accuracy_weight" json:"raw_accuracy_weight"`
	SyntaxAccuracyWeight float64 `datastore:"syntax_accuracy_weight" json:"syntax_accuracy_weight"`
	BackspacePenalty     float64 `datastore:"backspace_penalty" json:"backspace_penalty"`
	IdleTimePenalty      float64 `datastore:"idle_time_penalty" json:"idle_time_penalty"`
	ConsistencyBonus     float64 `datastore:"consistency_bonus" json:"consistency_bonus"`

	// Anti-cheat thresholds
	MaxKPS              float64 `datastore:"max_kps" json:"max_kps"`
	MinBurstConsistency float64 `datastore:"min_burst_consistency" json:"min_burst_consistency"`
	MaxErrorRate        float64 `datastore:"max_error_rate" json:"max_error_rate"`
	MinConfidenceScore  float64 `datastore:"min_confidence_score" json:"min_confidence_score"`

	// Language-specific adjustments
	LanguageMultipliers map[string]float64 `datastore:"language_multipliers" json:"language_multipliers"`
	ModeMultipliers     map[string]float64 `datastore:"mode_multipliers" json:"mode_multipliers"`

	// Metadata
	IsActive  bool      `datastore:"is_active" json:"is_active"`
	CreatedBy string    `datastore:"created_by" json:"created_by"`
	CreatedAt time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt time.Time `datastore:"updated_at" json:"updated_at"`
}

// ABTest entity for Datastore
// Manages A/B tests for scoring weights and UI variants
type ABTest struct {
	ID             string                   `datastore:"-" json:"id"`
	Name           string                   `datastore:"name" json:"name"`
	Description    string                   `datastore:"description" json:"description"`
	TestType       string                   `datastore:"test_type" json:"test_type"` // "scoring_weights", "ui_variant"
	Variants       []map[string]interface{} `datastore:"variants" json:"variants"`
	Status         string                   `datastore:"status" json:"status"` // "active", "paused", "completed"
	StartDate      time.Time                `datastore:"start_date" json:"start_date"`
	EndDate        time.Time                `datastore:"end_date" json:"end_date"`
	UserPercentage float64                  `datastore:"user_percentage" json:"user_percentage"` // Percentage of users to include
	ParticipantIDs []string                 `datastore:"participant_ids" json:"participant_ids"`
	Results        map[string]interface{}   `datastore:"results" json:"results"`
	CreatedBy      string                   `datastore:"created_by" json:"created_by"`
	CreatedAt      time.Time                `datastore:"created_at" json:"created_at"`
	UpdatedAt      time.Time                `datastore:"updated_at" json:"updated_at"`
}

// Assessment Models

// AssessmentBlueprint entity for Datastore
type AssessmentBlueprint struct {
	ID                string             `datastore:"-" json:"id"`
	Name              string             `datastore:"name" json:"name"`
	Description       string             `datastore:"description" json:"description"`
	Language          string             `datastore:"language" json:"language"`
	Difficulty        int                `datastore:"difficulty" json:"difficulty"`
	EstimatedDuration int                `datastore:"estimated_duration" json:"estimated_duration"` // minutes
	SnippetIDs        []string           `datastore:"snippet_ids" json:"snippet_ids"`
	PassingCriteria   AssessmentCriteria `datastore:"passing_criteria" json:"passing_criteria"`
	ScoringWeights    AssessmentWeights  `datastore:"scoring_weights" json:"scoring_weights"`
	Version           int                `datastore:"version" json:"version"`
	CreatedAt         time.Time          `datastore:"created_at" json:"created_at"`
}

// AssessmentCriteria embedded struct
type AssessmentCriteria struct {
	MinimumAccuracy          float64 `datastore:"minimum_accuracy" json:"minimum_accuracy"`
	MinimumSpeed             float64 `datastore:"minimum_speed" json:"minimum_speed"`
	MaximumErrorRate         float64 `datastore:"maximum_error_rate" json:"maximum_error_rate"`
	StructuralAccuracyWeight float64 `datastore:"structural_accuracy_weight" json:"structural_accuracy_weight"`
	SyntaxPenaltyMultiplier  float64 `datastore:"syntax_penalty_multiplier" json:"syntax_penalty_multiplier"`
	TimeLimit                *int    `datastore:"time_limit" json:"time_limit,omitempty"`
}

// AssessmentWeights embedded struct
type AssessmentWeights struct {
	Speed                float64 `datastore:"speed" json:"speed"`
	Accuracy             float64 `datastore:"accuracy" json:"accuracy"`
	StructuralConformity float64 `datastore:"structural_conformity" json:"structural_conformity"`
	SyntaxCorrectness    float64 `datastore:"syntax_correctness" json:"syntax_correctness"`
	Consistency          float64 `datastore:"consistency" json:"consistency"`
	ErrorRecovery        float64 `datastore:"error_recovery" json:"error_recovery"`
}

// AssessmentSession entity for Datastore
type AssessmentSession struct {
	ID                  string                    `datastore:"-" json:"id"`
	BlueprintID         string                    `datastore:"blueprint_id" json:"blueprint_id"`
	UserID              string                    `datastore:"user_id" json:"user_id"`
	StartedAt           time.Time                 `datastore:"started_at" json:"started_at"`
	CompletedAt         *time.Time                `datastore:"completed_at" json:"completed_at,omitempty"`
	Status              string                    `datastore:"status" json:"status"`
	CurrentSnippetIndex int                       `datastore:"current_snippet_index" json:"current_snippet_index"`
	SnippetResults      []AssessmentSnippetResult `datastore:"snippet_results" json:"snippet_results"`
	OverallResult       *AssessmentResult         `datastore:"overall_result" json:"overall_result,omitempty"`
	Metadata            AssessmentMetadata        `datastore:"metadata" json:"metadata"`
}

// AssessmentSnippetResult embedded struct
type AssessmentSnippetResult struct {
	SnippetID       string                    `datastore:"snippet_id" json:"snippet_id"`
	StartedAt       time.Time                 `datastore:"started_at" json:"started_at"`
	CompletedAt     *time.Time                `datastore:"completed_at" json:"completed_at,omitempty"`
	ExpectedText    string                    `datastore:"expected_text,noindex" json:"expected_text"`
	ActualText      string                    `datastore:"actual_text,noindex" json:"actual_text"`
	KeystrokeEvents []map[string]interface{}  `datastore:"keystroke_events,noindex" json:"keystroke_events"`
	Metrics         SessionMetrics            `datastore:"metrics" json:"metrics"`
	StructuralScore StructuralConformityScore `datastore:"structural_score" json:"structural_score"`
	Passed          bool                      `datastore:"passed" json:"passed"`
	TimeSpent       int64                     `datastore:"time_spent" json:"time_spent"` // milliseconds
}

// SessionMetrics embedded struct (reused from existing models)
type SessionMetrics struct {
	CPM               float64 `datastore:"cpm" json:"cpm"`
	TWPM              float64 `datastore:"twpm" json:"twpm"`
	KPS               float64 `datastore:"kps" json:"kps"`
	RawAccuracy       float64 `datastore:"raw_accuracy" json:"raw_accuracy"`
	TokenAccuracy     float64 `datastore:"token_accuracy" json:"token_accuracy"`
	SyntaxAccuracy    float64 `datastore:"syntax_accuracy" json:"syntax_accuracy"`
	BackspaceRate     float64 `datastore:"backspace_rate" json:"backspace_rate"`
	ErrorsPerMinute   float64 `datastore:"errors_per_minute" json:"errors_per_minute"`
	CorrectionLatency float64 `datastore:"correction_latency" json:"correction_latency"`
}

// StructuralConformityScore embedded struct
type StructuralConformityScore struct {
	ASTSimilarity         float64             `datastore:"ast_similarity" json:"ast_similarity"`
	TokenSequenceAccuracy float64             `datastore:"token_sequence_accuracy" json:"token_sequence_accuracy"`
	SyntaxValidationScore float64             `datastore:"syntax_validation_score" json:"syntax_validation_score"`
	StructuralPenalties   []StructuralPenalty `datastore:"structural_penalties" json:"structural_penalties"`
	OverallConformity     float64             `datastore:"overall_conformity" json:"overall_conformity"`
}

// StructuralPenalty embedded struct
type StructuralPenalty struct {
	Type          string  `datastore:"type" json:"type"`
	Severity      string  `datastore:"severity" json:"severity"`
	Position      int     `datastore:"position" json:"position"`
	Description   string  `datastore:"description" json:"description"`
	PenaltyPoints float64 `datastore:"penalty_points" json:"penalty_points"`
}

// AssessmentResult entity for Datastore
type AssessmentResult struct {
	SessionID                string                   `datastore:"-" json:"session_id"`
	OverallScore             int                      `datastore:"overall_score" json:"overall_score"`
	Passed                   bool                     `datastore:"passed" json:"passed"`
	Grade                    string                   `datastore:"grade" json:"grade"`
	Breakdown                AssessmentScoreBreakdown `datastore:"breakdown" json:"breakdown"`
	Recommendations          []string                 `datastore:"recommendations" json:"recommendations"`
	CertificateEligible      bool                     `datastore:"certificate_eligible" json:"certificate_eligible"`
	RetakeAllowed            bool                     `datastore:"retake_allowed" json:"retake_allowed"`
	NextAssessmentSuggestion *string                  `datastore:"next_assessment_suggestion" json:"next_assessment_suggestion,omitempty"`
	CreatedAt                time.Time                `datastore:"created_at" json:"created_at"`
}

// AssessmentScoreBreakdown embedded struct
type AssessmentScoreBreakdown struct {
	SpeedScore         float64 `datastore:"speed_score" json:"speed_score"`
	AccuracyScore      float64 `datastore:"accuracy_score" json:"accuracy_score"`
	StructuralScore    float64 `datastore:"structural_score" json:"structural_score"`
	SyntaxScore        float64 `datastore:"syntax_score" json:"syntax_score"`
	ConsistencyScore   float64 `datastore:"consistency_score" json:"consistency_score"`
	ErrorRecoveryScore float64 `datastore:"error_recovery_score" json:"error_recovery_score"`
	TotalPenalties     float64 `datastore:"total_penalties" json:"total_penalties"`
	BonusPoints        float64 `datastore:"bonus_points" json:"bonus_points"`
}

// AssessmentMetadata embedded struct
type AssessmentMetadata struct {
	Language          string                 `datastore:"language" json:"language"`
	Difficulty        int                    `datastore:"difficulty" json:"difficulty"`
	TotalSnippets     int                    `datastore:"total_snippets" json:"total_snippets"`
	EstimatedDuration int                    `datastore:"estimated_duration" json:"estimated_duration"`
	ActualDuration    *int                   `datastore:"actual_duration" json:"actual_duration,omitempty"`
	Environment       map[string]interface{} `datastore:"environment" json:"environment"`
	Settings          map[string]interface{} `datastore:"settings" json:"settings"`
}

// AssessmentSchedule entity for Datastore
type AssessmentSchedule struct {
	ID               string    `datastore:"-" json:"id"`
	UserID           string    `datastore:"user_id" json:"user_id"`
	AssessmentID     string    `datastore:"assessment_id" json:"assessment_id"`
	ScheduledAt      time.Time `datastore:"scheduled_at" json:"scheduled_at"`
	ReminderSent     bool      `datastore:"reminder_sent" json:"reminder_sent"`
	Completed        bool      `datastore:"completed" json:"completed"`
	RescheduledCount int       `datastore:"rescheduled_count" json:"rescheduled_count"`
}

// AssessmentBadge entity for Datastore
type AssessmentBadge struct {
	ID          string        `datastore:"-" json:"id"`
	Name        string        `datastore:"name" json:"name"`
	Description string        `datastore:"description" json:"description"`
	IconURL     string        `datastore:"icon_url" json:"icon_url"`
	Criteria    BadgeCriteria `datastore:"criteria" json:"criteria"`
	Rarity      string        `datastore:"rarity" json:"rarity"`
	EarnedAt    *time.Time    `datastore:"earned_at" json:"earned_at,omitempty"`
	CreatedAt   time.Time     `datastore:"created_at" json:"created_at"`
}

// BadgeCriteria embedded struct
type BadgeCriteria struct {
	MinimumScore      *int     `datastore:"minimum_score" json:"minimum_score,omitempty"`
	MinimumGrade      *string  `datastore:"minimum_grade" json:"minimum_grade,omitempty"`
	SpecificLanguage  *string  `datastore:"specific_language" json:"specific_language,omitempty"`
	ConsecutivePasses *int     `datastore:"consecutive_passes" json:"consecutive_passes,omitempty"`
	TimeConstraint    *int     `datastore:"time_constraint" json:"time_constraint,omitempty"`
	PerfectAccuracy   *bool    `datastore:"perfect_accuracy" json:"perfect_accuracy,omitempty"`
	SpeedThreshold    *float64 `datastore:"speed_threshold" json:"speed_threshold,omitempty"`
}

// UserBadge entity for Datastore (junction table for user-badge relationships)
type UserBadge struct {
	ID       string    `datastore:"-" json:"id"`
	UserID   string    `datastore:"user_id" json:"user_id"`
	BadgeID  string    `datastore:"badge_id" json:"badge_id"`
	EarnedAt time.Time `datastore:"earned_at" json:"earned_at"`
}

// AssessmentAnalytics struct for analytics responses
type AssessmentAnalytics struct {
	TotalAttempts          int                            `json:"total_attempts"`
	PassRate               float64                        `json:"pass_rate"`
	AverageScore           float64                        `json:"average_score"`
	AverageDuration        float64                        `json:"average_duration"`
	CommonFailurePoints    []FailurePoint                 `json:"common_failure_points"`
	DifficultyDistribution map[int]float64                `json:"difficulty_distribution"`
	LanguagePerformance    map[string]LanguagePerformance `json:"language_performance"`
}

// FailurePoint embedded struct
type FailurePoint struct {
	SnippetID           string  `json:"snippet_id"`
	Position            int     `json:"position"`
	ErrorType           string  `json:"error_type"`
	Frequency           float64 `json:"frequency"`
	AverageRecoveryTime int64   `json:"average_recovery_time"`
}

// LanguagePerformance embedded struct
type LanguagePerformance struct {
	AverageScore  float64  `json:"average_score"`
	PassRate      float64  `json:"pass_rate"`
	CommonErrors  []string `json:"common_errors"`
	StrengthAreas []string `json:"strength_areas"`
}

// RBAC Models for Role-Based Access Control

// Role entity for Datastore - defines roles in the system
type Role struct {
	ID          string    `datastore:"-" json:"id"`
	Name        string    `datastore:"name" json:"name"`                 // "admin", "moderator", "content_creator", "user"
	DisplayName string    `datastore:"display_name" json:"display_name"` // "Administrator", "Moderator", etc.
	Description string    `datastore:"description" json:"description"`
	Permissions []string  `datastore:"permissions" json:"permissions"` // List of permission IDs
	IsActive    bool      `datastore:"is_active" json:"is_active"`
	CreatedBy   string    `datastore:"created_by" json:"created_by"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

// Permission entity for Datastore - defines granular permissions
type Permission struct {
	ID          string    `datastore:"-" json:"id"`
	Name        string    `datastore:"name" json:"name"` // "create_content", "manage_users", etc.
	DisplayName string    `datastore:"display_name" json:"display_name"`
	Description string    `datastore:"description" json:"description"`
	Resource    string    `datastore:"resource" json:"resource"` // "content", "users", "sessions", etc.
	Action      string    `datastore:"action" json:"action"`     // "create", "read", "update", "delete"
	Scope       string    `datastore:"scope" json:"scope"`       // "own", "all", "team"
	IsActive    bool      `datastore:"is_active" json:"is_active"`
	CreatedBy   string    `datastore:"created_by" json:"created_by"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

// UserRole entity for Datastore - assigns roles to users
type UserRole struct {
	ID         string     `datastore:"-" json:"id"`
	UserID     string     `datastore:"user_id" json:"user_id"`
	RoleID     string     `datastore:"role_id" json:"role_id"`
	AssignedBy string     `datastore:"assigned_by" json:"assigned_by"`
	AssignedAt time.Time  `datastore:"assigned_at" json:"assigned_at"`
	IsActive   bool       `datastore:"is_active" json:"is_active"`
	ExpiresAt  *time.Time `datastore:"expires_at" json:"expires_at,omitempty"`
	Notes      string     `datastore:"notes" json:"notes"`
	CreatedAt  time.Time  `datastore:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `datastore:"updated_at" json:"updated_at"`
}

// Integration entity for Datastore
// Manages third-party service integrations (GitHub, VS Code, etc.)
type Integration struct {
	ID        string                 `datastore:"-" json:"id"`
	UserID    string                 `datastore:"user_id" json:"user_id"`
	Provider  string                 `datastore:"provider" json:"provider"` // "github", "vscode", "gitlab", "bitbucket"
	Status    string                 `datastore:"status" json:"status"`     // "active", "inactive", "error"
	Config    map[string]interface{} `datastore:"config" json:"config"`
	Metadata  map[string]interface{} `datastore:"metadata" json:"metadata"`
	CreatedAt time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt time.Time              `datastore:"updated_at" json:"updated_at"`
}

// Embed entity for Datastore
// Manages embedded typing practice sessions for workflow integration
type Embed struct {
	ID          string                 `datastore:"-" json:"id"`
	Platform    string                 `datastore:"platform" json:"platform"`         // "github-actions", "vscode", etc.
	ContentType string                 `datastore:"content_type" json:"content_type"` // "ci", "editor", "webhook"
	SessionID   string                 `datastore:"session_id" json:"session_id"`
	Config      map[string]interface{} `datastore:"config" json:"config"`
	UserID      string                 `datastore:"user_id" json:"user_id"`
	CreatedAt   time.Time              `datastore:"created_at" json:"created_at"`
	ExpiresAt   time.Time              `datastore:"expires_at" json:"expires_at"`
}

type SystemVersion struct {
	ID                  string    `datastore:"-" json:"id"`
	Version             string    `datastore:"version" json:"version"`
	Channel             string    `datastore:"channel" json:"channel"`
	Description         string    `datastore:"description" json:"description"`
	Changes             []string  `datastore:"changes" json:"changes"`
	BreakingChanges     []string  `datastore:"breaking_changes" json:"breaking_changes"`
	DeprecationNotices  []string  `datastore:"deprecation_notices" json:"deprecation_notices"`
	MigrationGuideURL   string    `datastore:"migration_guide_url" json:"migration_guide_url"`
	RequiresMigration   bool      `datastore:"requires_migration" json:"requires_migration"`
	RollbackTargetID    string    `datastore:"rollback_target_id" json:"rollback_target_id"`
	IsCurrent           bool      `datastore:"is_current" json:"is_current"`
	IsRollbackCandidate bool      `datastore:"is_rollback_candidate" json:"is_rollback_candidate"`
	ReleasedAt          time.Time `datastore:"released_at" json:"released_at"`
	CreatedAt           time.Time `datastore:"created_at" json:"created_at"`
	CreatedBy           string    `datastore:"created_by" json:"created_by"`
}

type ComplianceStandard struct {
	ID               string     `datastore:"-" json:"id"`
	Code             string     `datastore:"code" json:"code"`
	Name             string     `datastore:"name" json:"name"`
	Version          string     `datastore:"version" json:"version"`
	Jurisdiction     string     `datastore:"jurisdiction" json:"jurisdiction"`
	Status           string     `datastore:"status" json:"status"`
	CoverageAreas    []string   `datastore:"coverage_areas" json:"coverage_areas"`
	DocumentationURL string     `datastore:"documentation_url" json:"documentation_url"`
	Notes            string     `datastore:"notes" json:"notes"`
	EffectiveFrom    time.Time  `datastore:"effective_from" json:"effective_from"`
	EffectiveTo      *time.Time `datastore:"effective_to" json:"effective_to,omitempty"`
	CreatedAt        time.Time  `datastore:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `datastore:"updated_at" json:"updated_at"`
	CreatedBy        string     `datastore:"created_by" json:"created_by"`
	UpdatedBy        string     `datastore:"updated_by" json:"updated_by"`
}
