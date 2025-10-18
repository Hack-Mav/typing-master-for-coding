package models

import (
	"time"

	"cloud.google.com/go/datastore"
)

// User entity for Datastore
type User struct {
	ID             string                 `datastore:"-" json:"id"`
	Handle         string                 `datastore:"handle" json:"handle"`
	Email          string                 `datastore:"email" json:"email"`
	Locale         string                 `datastore:"locale" json:"locale"`
	KeyboardLayout string                 `datastore:"keyboard_layout" json:"keyboard_layout"`
	PrivacyMode    bool                   `datastore:"privacy_mode" json:"privacy_mode"`
	Settings       map[string]interface{} `datastore:"settings" json:"settings"`
	CreatedAt      time.Time              `datastore:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `datastore:"updated_at" json:"updated_at"`
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
	Version         string                 `datastore:"version" json:"version"`
	ParserID        string                 `datastore:"parser_id" json:"parser_id"`
	GrammarConfig   map[string]interface{} `datastore:"grammar_config" json:"grammar_config"`
	WhitespaceRules map[string]interface{} `datastore:"whitespace_rules" json:"whitespace_rules"`
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

func (da *DifficultyAssessment) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(da, ps)
}