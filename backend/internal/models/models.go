package models

import (
	"time"

	"cloud.google.com/go/datastore"
)

// User entity for Datastore
type User struct {
	ID                  string                 `datastore:"-" json:"id"`
	Handle              string                 `datastore:"handle" json:"handle"`
	Email               string                 `datastore:"email" json:"email"`
	PasswordHash        string                 `datastore:"password_hash" json:"-"` // Never expose in JSON
	Role                string                 `datastore:"role" json:"role"` // "user", "admin", "moderator"
	IsAnonymous         bool                   `datastore:"is_anonymous" json:"is_anonymous"`
	Locale              string                 `datastore:"locale" json:"locale"`
	KeyboardLayout      string                 `datastore:"keyboard_layout" json:"keyboard_layout"`
	PrivacyMode         bool                   `datastore:"privacy_mode" json:"privacy_mode"`
	TelemetryConsent    bool                   `datastore:"telemetry_consent" json:"telemetry_consent"`
	DataProcessingConsent bool                 `datastore:"data_processing_consent" json:"data_processing_consent"`
	Settings            map[string]interface{} `datastore:"settings" json:"settings"`
	CreatedAt           time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt           time.Time              `datastore:"updated_at" json:"updated_at"`
	LastLoginAt         *time.Time             `datastore:"last_login_at" json:"last_login_at,omitempty"`
}

// LoadKey implements the PropertyLoadSaver interface
func (u *User) LoadKey(k *datastore.Key) error {
	u.ID = k.Name
	if u.ID == "" && k.ID != 0 {
		u.ID = k.Encode()
	}
	return nil
}

// Save implements the PropertyLoadSaver interface
func (u *User) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(u)
}

// Load implements the PropertyLoadSaver interface
func (u *User) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(u, ps)
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

func (l *Language) LoadKey(k *datastore.Key) error {
	l.ID = k.Name
	return nil
}

func (l *Language) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(l)
}

func (l *Language) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(l, ps)
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

func (l *Lesson) LoadKey(k *datastore.Key) error {
	l.ID = k.Name
	if l.ID == "" && k.ID != 0 {
		l.ID = k.Encode()
	}
	return nil
}

func (l *Lesson) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(l)
}

func (l *Lesson) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(l, ps)
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

func (s *Snippet) LoadKey(k *datastore.Key) error {
	s.ID = k.Name
	if s.ID == "" && k.ID != 0 {
		s.ID = k.Encode()
	}
	return nil
}

func (s *Snippet) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(s)
}

func (s *Snippet) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(s, ps)
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

func (s *Session) LoadKey(k *datastore.Key) error {
	s.ID = k.Name
	if s.ID == "" && k.ID != 0 {
		s.ID = k.Encode()
	}
	return nil
}

func (s *Session) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(s)
}

func (s *Session) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(s, ps)
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
	Metadata       map[string]interface{} `datastore:"metadata" json:"metadata"`
	CreatedAt      time.Time              `datastore:"created_at" json:"created_at"`
}

func (se *SessionEvent) LoadKey(k *datastore.Key) error {
	se.ID = k.Name
	if se.ID == "" && k.ID != 0 {
		se.ID = k.Encode()
	}
	return nil
}

func (se *SessionEvent) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(se)
}

func (se *SessionEvent) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(se, ps)
}

// Result entity for Datastore
type Result struct {
	SessionID       string                 `datastore:"-" json:"session_id"`
	CPM             float64                `datastore:"cpm" json:"cpm"`
	TWPM            float64                `datastore:"twpm" json:"twpm"`
	RawAccuracy     float64                `datastore:"raw_accuracy" json:"raw_accuracy"`
	TokenAccuracy   float64                `datastore:"token_accuracy" json:"token_accuracy"`
	SyntaxAccuracy  float64                `datastore:"syntax_accuracy" json:"syntax_accuracy"`
	BackspaceRate   float64                `datastore:"backspace_rate" json:"backspace_rate"`
	CompositeScore  int                    `datastore:"composite_score" json:"composite_score"`
	Breakdown       map[string]interface{} `datastore:"breakdown" json:"breakdown"`
	CreatedAt       time.Time              `datastore:"created_at" json:"created_at"`
}

func (r *Result) LoadKey(k *datastore.Key) error {
	r.SessionID = k.Name
	return nil
}

func (r *Result) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(r)
}

func (r *Result) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(r, ps)
}

// Playlist entity for Datastore
// Playlists are collections of lessons or snippets for custom practice
type Playlist struct {
	ID          string   `datastore:"-" json:"id"`
	UserID      string   `datastore:"user_id" json:"user_id"`
	Title       string   `datastore:"title" json:"title"`
	Description string   `datastore:"description" json:"description"`
	LanguageID  string   `datastore:"language_id" json:"language_id"`
	LessonIDs   []string `datastore:"lesson_ids" json:"lesson_ids"`
	SnippetIDs  []string `datastore:"snippet_ids" json:"snippet_ids"`
	Tags        []string `datastore:"tags" json:"tags"`
	IsPublic    bool     `datastore:"is_public" json:"is_public"`
	Difficulty  int      `datastore:"difficulty" json:"difficulty"`
	Version     int      `datastore:"version" json:"version"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

func (p *Playlist) LoadKey(k *datastore.Key) error {
	p.ID = k.Name
	if p.ID == "" && k.ID != 0 {
		p.ID = k.Encode()
	}
	return nil
}

func (p *Playlist) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(p)
}

func (p *Playlist) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(p, ps)
}

// ContentVersion entity for Datastore
// Tracks versions of content for audit and rollback capabilities
type ContentVersion struct {
	ID            string                 `datastore:"-" json:"id"`
	ContentType   string                 `datastore:"content_type" json:"content_type"` // "language", "lesson", "snippet", "playlist"
	ContentID     string                 `datastore:"content_id" json:"content_id"`
	Version       int                    `datastore:"version" json:"version"`
	Content       map[string]interface{} `datastore:"content" json:"content"`
	Checksum      string                 `datastore:"checksum" json:"checksum"`
	CreatedBy     string                 `datastore:"created_by" json:"created_by"`
	ChangeNotes   string                 `datastore:"change_notes" json:"change_notes"`
	IsActive      bool                   `datastore:"is_active" json:"is_active"`
	CreatedAt     time.Time              `datastore:"created_at" json:"created_at"`
}

func (cv *ContentVersion) LoadKey(k *datastore.Key) error {
	cv.ID = k.Name
	if cv.ID == "" && k.ID != 0 {
		cv.ID = k.Encode()
	}
	return nil
}

func (cv *ContentVersion) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(cv)
}

func (cv *ContentVersion) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(cv, ps)
}

// ContentValidation entity for Datastore
// Stores validation results and checksums for content integrity
type ContentValidation struct {
	ID            string    `datastore:"-" json:"id"`
	ContentType   string    `datastore:"content_type" json:"content_type"` // "language", "lesson", "snippet", "playlist"
	ContentID     string    `datastore:"content_id" json:"content_id"`
	Checksum      string    `datastore:"checksum" json:"checksum"`
	ValidationStatus string `datastore:"validation_status" json:"validation_status"` // "valid", "invalid", "warning"
	ValidationErrors []string `datastore:"validation_errors" json:"validation_errors"`
	ValidatedAt   time.Time `datastore:"validated_at" json:"validated_at"`
}

func (cv *ContentValidation) LoadKey(k *datastore.Key) error {
	cv.ID = k.Name
	if cv.ID == "" && k.ID != 0 {
		cv.ID = k.Encode()
	}
	return nil
}

func (cv *ContentValidation) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(cv)
}

func (cv *ContentValidation) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(cv, ps)
}

// LessonProgress entity for Datastore
// Tracks user progress through lessons and lesson flows
type LessonProgress struct {
	ID               string    `datastore:"-" json:"id"`
	UserID           string    `datastore:"user_id" json:"user_id"`
	LessonID         string    `datastore:"lesson_id" json:"lesson_id"`
	CurrentStage     string    `datastore:"current_stage" json:"current_stage"` // "intro", "core", "idioms", "advanced", "review"
	CompletedStages  []string  `datastore:"completed_stages" json:"completed_stages"`
	ProgressPercent  float64   `datastore:"progress_percent" json:"progress_percent"`
	TokensCovered    []string  `datastore:"tokens_covered" json:"tokens_covered"`
	IsUnlocked       bool      `datastore:"is_unlocked" json:"is_unlocked"`
	IsCompleted      bool      `datastore:"is_completed" json:"is_completed"`
	BestScore        int       `datastore:"best_score" json:"best_score"`
	AttemptsCount    int       `datastore:"attempts_count" json:"attempts_count"`
	LastAttemptAt    time.Time `datastore:"last_attempt_at" json:"last_attempt_at"`
	CreatedAt        time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt        time.Time `datastore:"updated_at" json:"updated_at"`
}

func (lp *LessonProgress) LoadKey(k *datastore.Key) error {
	lp.ID = k.Name
	if lp.ID == "" && k.ID != 0 {
		lp.ID = k.Encode()
	}
	return nil
}

func (lp *LessonProgress) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(lp)
}

func (lp *LessonProgress) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(lp, ps)
}

// DifficultyAssessment entity for Datastore
// Stores difficulty assessments and adaptive progression data
type DifficultyAssessment struct {
	ID                  string                 `datastore:"-" json:"id"`
	ContentType         string                 `datastore:"content_type" json:"content_type"` // "lesson", "snippet"
	ContentID           string                 `datastore:"content_id" json:"content_id"`
	AssessedDifficulty  int                    `datastore:"assessed_difficulty" json:"assessed_difficulty"`
	ComplexityScore     float64                `datastore:"complexity_score" json:"complexity_score"`
	TokenComplexity     map[string]float64     `datastore:"token_complexity" json:"token_complexity"`
	EstimatedTime       int                    `datastore:"estimated_time" json:"estimated_time"` // minutes
	SuccessRate         float64                `datastore:"success_rate" json:"success_rate"`
	AverageAttempts      float64                `datastore:"average_attempts" json:"average_attempts"`
	AdaptiveLevel       int                    `datastore:"adaptive_level" json:"adaptive_level"`
	AssessmentMetadata  map[string]interface{} `datastore:"assessment_metadata" json:"assessment_metadata"`
	AssessedAt          time.Time              `datastore:"assessed_at" json:"assessed_at"`
}

func (da *DifficultyAssessment) LoadKey(k *datastore.Key) error {
	da.ID = k.Name
	if da.ID == "" && k.ID != 0 {
		da.ID = k.Encode()
	}
	return nil
}

func (da *DifficultyAssessment) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(da)
}

// ScoringMetrics entity for Datastore (extends Result with more detailed metrics)
type ScoringMetrics struct {
	ID              string                 `datastore:"-" json:"id"`
	SessionID       string                 `datastore:"session_id" json:"session_id"`
	UserID          string                 `datastore:"user_id" json:"user_id"`
	LanguageID      string                 `datastore:"language_id" json:"language_id"`
	Mode            string                 `datastore:"mode" json:"mode"`
	
	// Speed metrics
	CPM             float64                `datastore:"cpm" json:"cpm"`
	TWPM            float64                `datastore:"twpm" json:"twpm"`
	KPS             float64                `datastore:"kps" json:"kps"`
	
	// Accuracy metrics
	RawAccuracy     float64                `datastore:"raw_accuracy" json:"raw_accuracy"`
	TokenAccuracy   float64                `datastore:"token_accuracy" json:"token_accuracy"`
	SyntaxAccuracy  float64                `datastore:"syntax_accuracy" json:"syntax_accuracy"`
	WhitespaceAccuracy float64             `datastore:"whitespace_accuracy" json:"whitespace_accuracy"`
	
	// Efficiency metrics
	BackspaceRate   float64                `datastore:"backspace_rate" json:"backspace_rate"`
	CorrectionRate  float64                `datastore:"correction_rate" json:"correction_rate"`
	IdleTimePercent float64                `datastore:"idle_time_percent" json:"idle_time_percent"`
	
	// Composite scores
	CompositeScore  int                    `datastore:"composite_score" json:"composite_score"`
	ConsistencyScore float64               `datastore:"consistency_score" json:"consistency_score"`
	EfficiencyScore float64                `datastore:"efficiency_score" json:"efficiency_score"`
	
	// Detailed breakdown
	ErrorClusters   map[string]interface{} `datastore:"error_clusters" json:"error_clusters"`
	PerformanceInsights map[string]interface{} `datastore:"performance_insights" json:"performance_insights"`
	TypingPatterns  map[string]interface{} `datastore:"typing_patterns" json:"typing_patterns"`
	
	// Anti-cheat flags
	SuspiciousActivity bool                 `datastore:"suspicious_activity" json:"suspicious_activity"`
	CheatFlags      []string               `datastore:"cheat_flags" json:"cheat_flags"`
	ConfidenceScore float64                `datastore:"confidence_score" json:"confidence_score"`
	
	// Timestamps
	CalculatedAt    time.Time              `datastore:"calculated_at" json:"calculated_at"`
	CreatedAt       time.Time              `datastore:"created_at" json:"created_at"`
}

func (sm *ScoringMetrics) LoadKey(k *datastore.Key) error {
	sm.ID = k.Name
	if sm.ID == "" && k.ID != 0 {
		sm.ID = k.Encode()
	}
	return nil
}

func (sm *ScoringMetrics) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(sm)
}

// Leaderboard entity for Datastore
type Leaderboard struct {
	ID              string                 `datastore:"-" json:"id"`
	UserID          string                 `datastore:"user_id" json:"user_id"`
	SessionID       string                 `datastore:"session_id" json:"session_id"`
	LanguageID      string                 `datastore:"language_id" json:"language_id"`
	Mode            string                 `datastore:"mode" json:"mode"`
	Scope           string                 `datastore:"scope" json:"scope"` // "global", "friends", "organization"
	TimeWindow      string                 `datastore:"time_window" json:"time_window"` // "daily", "weekly", "monthly", "all_time"
	
	// Rankings
	Rank            int                    `datastore:"rank" json:"rank"`
	Score           int                    `datastore:"score" json:"score"`
	CPM             float64                `datastore:"cpm" json:"cpm"`
	TWPM            float64                `datastore:"twpm" json:"twpm"`
	Accuracy        float64                `datastore:"accuracy" json:"accuracy"`
	
	// Metadata
	MetricsSnapshot map[string]interface{} `datastore:"metrics_snapshot" json:"metrics_snapshot"`
	Badge           string                 `datastore:"badge" json:"badge"`
	IsVerified      bool                   `datastore:"is_verified" json:"is_verified"`
	VerificationMethod string               `datastore:"verification_method" json:"verification_method"`
	
	// Timestamps
	RecordedAt      time.Time              `datastore:"recorded_at" json:"recorded_at"`
	WindowStart     time.Time              `datastore:"window_start" json:"window_start"`
	WindowEnd       time.Time              `datastore:"window_end" json:"window_end"`
}

func (l *Leaderboard) LoadKey(k *datastore.Key) error {
	l.ID = k.Name
	if l.ID == "" && k.ID != 0 {
		l.ID = k.Encode()
	}
	return nil
}

func (l *Leaderboard) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(l)
}

// AntiCheatReport entity for Datastore
type AntiCheatReport struct {
	ID                  string                 `datastore:"-" json:"id"`
	SessionID           string                 `datastore:"session_id" json:"session_id"`
	UserID              string                 `datastore:"user_id" json:"user_id"`
	
	// Detection results
	PasteDetected       bool                   `datastore:"paste_detected" json:"paste_detected"`
	UnrealisticKPS      bool                   `datastore:"unrealistic_kps" json:"unrealistic_kps"`
	AutoTypePattern     bool                   `datastore:"auto_type_pattern" json:"auto_type_pattern"`
	WindowFocusLost     bool                   `datastore:"window_focus_lost" json:"window_focus_lost"`
	AnomalousTiming     bool                   `datastore:"anomalous_timing" json:"anomalous_timing"`
	
	// Analysis data
	KPSVariance         float64                `datastore:"kps_variance" json:"kps_variance"`
	BurstConsistency    float64                `datastore:"burst_consistency" json:"burst_consistency"`
	ErrorPatternScore   float64                `datastore:"error_pattern_score" json:"error_pattern_score"`
	StatisticalAnomaly  float64                `datastore:"statistical_anomaly" json:"statistical_anomaly"`
	
	// Evidence
	SuspiciousEvents    []string               `datastore:"suspicious_events" json:"suspicious_events"`
	EvidenceDetails     map[string]interface{} `datastore:"evidence_details" json:"evidence_details"`
	ConfidenceScore     float64                `datastore:"confidence_score" json:"confidence_score"`
	RiskLevel           string                 `datastore:"risk_level" json:"risk_level"` // "low", "medium", "high", "critical"
	
	// Actions taken
	FlaggedForReview    bool                   `datastore:"flagged_for_review" json:"flagged_for_review"`
	ScoreInvalidated    bool                   `datastore:"score_invalidated" json:"score_invalidated"`
	LeaderboardExcluded bool                   `datastore:"leaderboard_excluded" json:"leaderboard_excluded"`
	
	// Timestamps
	DetectedAt          time.Time              `datastore:"detected_at" json:"detected_at"`
	ReviewedAt          *time.Time             `datastore:"reviewed_at" json:"reviewed_at"`
	ReviewedBy          string                 `datastore:"reviewed_by" json:"reviewed_by"`
}

func (acr *AntiCheatReport) LoadKey(k *datastore.Key) error {
	acr.ID = k.Name
	if acr.ID == "" && k.ID != 0 {
		acr.ID = k.Encode()
	}
	return nil
}

func (acr *AntiCheatReport) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(acr)
}

// Tournament entity for Datastore
type Tournament struct {
	ID                  string                 `datastore:"-" json:"id"`
	Title               string                 `datastore:"title" json:"title"`
	Description         string                 `datastore:"description" json:"description"`
	LanguageID          string                 `datastore:"language_id" json:"language_id"`
	Mode                string                 `datastore:"mode" json:"mode"`
	
	// Tournament settings
	VerificationRequired bool                   `datastore:"verification_required" json:"verification_required"`
	VerificationMethod  string                 `datastore:"verification_method" json:"verification_method"` // "webcam", "hid", "both"
	AntiCheatLevel      string                 `datastore:"anti_cheat_level" json:"anti_cheat_level"` // "standard", "enhanced", "maximum"
	
	// Schedule
	RegistrationStart   time.Time              `datastore:"registration_start" json:"registration_start"`
	RegistrationEnd     time.Time              `datastore:"registration_end" json:"registration_end"`
	StartTime           time.Time              `datastore:"start_time" json:"start_time"`
	EndTime             time.Time              `datastore:"end_time" json:"end_time"`
	DurationMinutes     int                    `datastore:"duration_minutes" json:"duration_minutes"`
	
	// Participation
	MaxParticipants     int                    `datastore:"max_participants" json:"max_participants"`
	CurrentParticipants int                    `datastore:"current_participants" json:"current_participants"`
	ParticipantIDs      []string               `datastore:"participant_ids" json:"participant_ids"`
	
	// Prizes and rewards
	PrizePool           float64                `datastore:"prize_pool" json:"prize_pool"`
	PrizeDistribution   map[string]float64     `datastore:"prize_distribution" json:"prize_distribution"`
	BadgeReward         string                 `datastore:"badge_reward" json:"badge_reward"`
	
	// Status
	Status              string                 `datastore:"status" json:"status"` // "upcoming", "registration", "active", "completed", "cancelled"
	
	// Results
	FinalRankings       []string               `datastore:"final_rankings" json:"final_rankings"`
	Winners             []string               `datastore:"winners" json:"winners"`
	
	// Metadata
	CreatedBy           string                 `datastore:"created_by" json:"created_by"`
	CreatedAt           time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt           time.Time              `datastore:"updated_at" json:"updated_at"`
}

func (t *Tournament) LoadKey(k *datastore.Key) error {
	t.ID = k.Name
	if t.ID == "" && k.ID != 0 {
		t.ID = k.Encode()
	}
	return nil
}

func (t *Tournament) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(t)
}

func (t *Tournament) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(t, ps)
}

// TournamentParticipant entity for Datastore
type TournamentParticipant struct {
	ID                string                 `datastore:"-" json:"id"`
	TournamentID      string                 `datastore:"tournament_id" json:"tournament_id"`
	UserID            string                 `datastore:"user_id" json:"user_id"`
	
	// Registration info
	RegisteredAt      time.Time              `datastore:"registered_at" json:"registered_at"`
	VerificationStatus string                `datastore:"verification_status" json:"verification_status"` // "pending", "verified", "failed"
	VerificationData  map[string]interface{} `datastore:"verification_data" json:"verification_data"`
	
	// Tournament results
	SessionID         string                 `datastore:"session_id" json:"session_id"`
	Score             int                    `datastore:"score" json:"score"`
	Rank              int                    `datastore:"rank" json:"rank"`
	CompletedAt       *time.Time             `datastore:"completed_at" json:"completed_at"`
	
	// Status
	Status            string                 `datastore:"status" json:"status"` // "registered", "active", "completed", "disqualified"
	DisqualificationReason string              `datastore:"disqualification_reason" json:"disqualification_reason"`
	
	CreatedAt         time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt         time.Time              `datastore:"updated_at" json:"updated_at"`
}

func (tp *TournamentParticipant) LoadKey(k *datastore.Key) error {
	tp.ID = k.Name
	if tp.ID == "" && k.ID != 0 {
		tp.ID = k.Encode()
	}
	return nil
}

func (tp *TournamentParticipant) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(tp)
}

// ScoringConfig entity for Datastore
// Stores configurable scoring weights and parameters
type ScoringConfig struct {
	ID                  string                 `datastore:"-" json:"id"`
	Version             string                 `datastore:"version" json:"version"`
	
	// Weight configurations
	TWPMWeight          float64                `datastore:"twpm_weight" json:"twpm_weight"`
	RawAccuracyWeight   float64                `datastore:"raw_accuracy_weight" json:"raw_accuracy_weight"`
	SyntaxAccuracyWeight float64               `datastore:"syntax_accuracy_weight" json:"syntax_accuracy_weight"`
	BackspacePenalty    float64                `datastore:"backspace_penalty" json:"backspace_penalty"`
	IdleTimePenalty     float64                `datastore:"idle_time_penalty" json:"idle_time_penalty"`
	ConsistencyBonus    float64                `datastore:"consistency_bonus" json:"consistency_bonus"`
	
	// Anti-cheat thresholds
	MaxKPS              float64                `datastore:"max_kps" json:"max_kps"`
	MinBurstConsistency float64                `datastore:"min_burst_consistency" json:"min_burst_consistency"`
	MaxErrorRate        float64                `datastore:"max_error_rate" json:"max_error_rate"`
	MinConfidenceScore  float64                `datastore:"min_confidence_score" json:"min_confidence_score"`
	
	// Language-specific adjustments
	LanguageMultipliers map[string]float64     `datastore:"language_multipliers" json:"language_multipliers"`
	ModeMultipliers     map[string]float64     `datastore:"mode_multipliers" json:"mode_multipliers"`
	
	// Metadata
	IsActive            bool                   `datastore:"is_active" json:"is_active"`
	CreatedBy           string                 `datastore:"created_by" json:"created_by"`
	CreatedAt           time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt           time.Time              `datastore:"updated_at" json:"updated_at"`
}

func (sc *ScoringConfig) LoadKey(k *datastore.Key) error {
	sc.ID = k.Name
	if sc.ID == "" && k.ID != 0 {
		sc.ID = k.Encode()
	}
	return nil
}

func (sc *ScoringConfig) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(sc)
}

// ABTest entity for Datastore
// Manages A/B tests for scoring weights and UI variants
type ABTest struct {
	ID             string                 `datastore:"-" json:"id"`
	Name           string                 `datastore:"name" json:"name"`
	Description    string                 `datastore:"description" json:"description"`
	TestType       string                 `datastore:"test_type" json:"test_type"` // "scoring_weights", "ui_variant"
	Variants       []map[string]interface{} `datastore:"variants" json:"variants"`
	Status         string                 `datastore:"status" json:"status"` // "active", "paused", "completed"
	StartDate      time.Time              `datastore:"start_date" json:"start_date"`
	EndDate        time.Time              `datastore:"end_date" json:"end_date"`
	UserPercentage float64                `datastore:"user_percentage" json:"user_percentage"` // Percentage of users to include
	ParticipantIDs []string               `datastore:"participant_ids" json:"participant_ids"`
	Results        map[string]interface{} `datastore:"results" json:"results"`
	CreatedBy      string                 `datastore:"created_by" json:"created_by"`
	CreatedAt      time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `datastore:"updated_at" json:"updated_at"`
}

func (abt *ABTest) LoadKey(k *datastore.Key) error {
	abt.ID = k.Name
	if abt.ID == "" && k.ID != 0 {
		abt.ID = k.Encode()
	}
	return nil
}

func (abt *ABTest) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(abt)
}

func (abt *ABTest) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(abt, ps)
}