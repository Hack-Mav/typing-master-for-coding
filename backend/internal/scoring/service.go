package scoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
)

// Service handles scoring computation and metrics processing
type Service struct {
	dsClient    *datastore.Client
	cache       *sync.Map
	config      *ScoringConfig
	eventQueue  chan *ScoringEvent
	workerCount int
}

// ScoringEvent represents a typing event for processing
type ScoringEvent struct {
	SessionID     string                 `json:"session_id"`
	UserID        string                 `json:"user_id"`
	EventType     string                 `json:"event_type"` // "keystroke", "session_end"
	Timestamp     int64                  `json:"timestamp"`
	KeyPressed    string                 `json:"key_pressed,omitempty"`
	Action        string                 `json:"action,omitempty"` // "down", "up"
	CursorPos     int                    `json:"cursor_pos"`
	ErrorFlag     bool                   `json:"error_flag"`
	ExpectedToken string                 `json:"expected_token,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SessionMetrics represents computed metrics for a session
type SessionMetrics struct {
	SessionID           string                 `json:"session_id"`
	UserID              string                 `json:"user_id"`
	LanguageID          string                 `json:"language_id"`
	Mode                string                 `json:"mode"`
	DurationMs          int64                  `json:"duration_ms"`

	// Speed metrics
	CPM                 float64                `json:"cpm"`
	TWPM                float64                `json:"twpm"`
	KPS                 float64                `json:"kps"`

	// Accuracy metrics
	RawAccuracy         float64                `json:"raw_accuracy"`
	TokenAccuracy       float64                `json:"token_accuracy"`
	SyntaxAccuracy      float64                `json:"syntax_accuracy"`
	WhitespaceAccuracy  float64                `json:"whitespace_accuracy"`

	// Efficiency metrics
	BackspaceRate       float64                `json:"backspace_rate"`
	CorrectionRate      float64                `json:"correction_rate"`
	IdleTimePercent     float64                `json:"idle_time_percent"`

	// Composite scores
	CompositeScore      int                    `json:"composite_score"`
	ConsistencyScore    float64                `json:"consistency_score"`
	EfficiencyScore     float64                `json:"efficiency_score"`

	// Detailed analysis
	ErrorClusters       map[string]interface{} `json:"error_clusters"`
	PerformanceInsights map[string]interface{} `json:"performance_insights"`
	TypingPatterns      map[string]interface{} `json:"typing_patterns"`

	// Anti-cheat analysis
	SuspiciousActivity  bool                   `json:"suspicious_activity"`
	CheatFlags          []string               `json:"cheat_flags"`
	ConfidenceScore     float64                `json:"confidence_score"`

	CalculatedAt        time.Time              `json:"calculated_at"`
}

// ScoringConfig holds configurable scoring parameters
type ScoringConfig struct {
	Version             string             `json:"version"`
	TWPMWeight          float64            `json:"twpm_weight"`
	RawAccuracyWeight   float64            `json:"raw_accuracy_weight"`
	SyntaxAccuracyWeight float64           `json:"syntax_accuracy_weight"`
	BackspacePenalty    float64            `json:"backspace_penalty"`
	IdleTimePenalty     float64            `json:"idle_time_penalty"`
	ConsistencyBonus    float64            `json:"consistency_bonus"`
	MaxKPS              float64            `json:"max_kps"`
	MinBurstConsistency float64            `json:"min_burst_consistency"`
	MaxErrorRate        float64            `json:"max_error_rate"`
	MinConfidenceScore  float64            `json:"min_confidence_score"`
	LanguageMultipliers map[string]float64 `json:"language_multipliers"`
	ModeMultipliers     map[string]float64 `json:"mode_multipliers"`
}

// NewService creates a new scoring service instance
func NewService(dsClient *datastore.Client) *Service {
	service := &Service{
		dsClient:    dsClient,
		cache:       &sync.Map{},
		eventQueue:  make(chan *ScoringEvent, 10000), // Buffer for 10k events
		workerCount: 4,
	}

	// Load default configuration
	service.config = service.getDefaultConfig()

	// Start event processing workers
	service.startWorkers()

	return service
}

// ProcessEvent processes a single scoring event
func (s *Service) ProcessEvent(ctx context.Context, event *ScoringEvent) error {
	select {
	case s.eventQueue <- event:
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// ProcessSession processes all events for a session and computes final metrics
func (s *Service) ProcessSession(ctx context.Context, sessionID string) (*SessionMetrics, error) {
	// Get all events for the session
	events, err := s.getSessionEvents(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for session %s", sessionID)
	}

	// Compute metrics from events
	metrics := s.computeMetrics(events)

	// Apply anti-cheat analysis
	antiCheatReport := s.analyzeAntiCheat(events)
	metrics.SuspiciousActivity = antiCheatReport.SuspiciousActivity
	metrics.CheatFlags = antiCheatReport.CheatFlags
	metrics.ConfidenceScore = antiCheatReport.ConfidenceScore

	// Store metrics
	err = s.storeMetrics(ctx, metrics)
	if err != nil {
		log.Printf("Failed to store metrics: %v", err)
		// Don't fail the entire operation for storage issues
	}

	// Store anti-cheat report if suspicious
	if antiCheatReport.SuspiciousActivity {
		err = s.storeAntiCheatReport(ctx, antiCheatReport)
		if err != nil {
			log.Printf("Failed to store anti-cheat report: %v", err)
		}
	}

	return metrics, nil
}

// GetLeaderboard retrieves leaderboard rankings for given parameters
func (s *Service) GetLeaderboard(ctx context.Context, languageID, mode, scope, timeWindow string, limit int) ([]LeaderboardEntry, error) {
	cacheKey := fmt.Sprintf("leaderboard:%s:%s:%s:%s:%d", languageID, mode, scope, timeWindow, limit)

	// Check cache first
	if cached, ok := s.cache.Load(cacheKey); ok {
		if entries, ok := cached.([]LeaderboardEntry); ok {
			return entries, nil
		}
	}

	// Query leaderboard data
	entries, err := s.queryLeaderboard(ctx, languageID, mode, scope, timeWindow, limit)
	if err != nil {
		return nil, err
	}

	// Cache results
	s.cache.Store(cacheKey, entries)

	return entries, nil
}

// UpdateLeaderboard updates leaderboard rankings for a completed session
func (s *Service) UpdateLeaderboard(ctx context.Context, sessionID string) error {
	// Get session metrics
	metrics, err := s.getStoredMetrics(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	if metrics == nil {
		return fmt.Errorf("no metrics found for session %s", sessionID)
	}

	// Check if session passed anti-cheat validation
	if metrics.SuspiciousActivity && metrics.ConfidenceScore < s.config.MinConfidenceScore {
		log.Printf("Session %s flagged for anti-cheat review, skipping leaderboard update", sessionID)
		return nil
	}

	// Update leaderboards for different scopes and time windows
	scopes := []string{"global", "friends"}
	timeWindows := []string{"daily", "weekly", "monthly", "all_time"}

	for _, scope := range scopes {
		for _, timeWindow := range timeWindows {
			err = s.updateLeaderboardEntry(ctx, metrics, scope, timeWindow)
			if err != nil {
				log.Printf("Failed to update leaderboard %s/%s: %v", scope, timeWindow, err)
				// Continue with other updates
			}
		}
	}

	return nil
}

// startWorkers starts background workers for processing events
func (s *Service) startWorkers() {
	for i := 0; i < s.workerCount; i++ {
		go s.eventWorker()
	}
}

// eventWorker processes events from the queue
func (s *Service) eventWorker() {
	for event := range s.eventQueue {
		ctx := context.Background()
		s.processEventBatch(ctx, []*ScoringEvent{event})
	}
}

// processEventBatch processes a batch of events
func (s *Service) processEventBatch(ctx context.Context, events []*ScoringEvent) {
	// Group events by session
	sessionEvents := make(map[string][]*ScoringEvent)
	for _, event := range events {
		sessionEvents[event.SessionID] = append(sessionEvents[event.SessionID], event)
	}

	// Process each session
	for sessionID, sessionEventList := range sessionEvents {
		// Check if session is complete
		if s.isSessionComplete(sessionEventList) {
			_, err := s.ProcessSession(ctx, sessionID)
			if err != nil {
				log.Printf("Failed to process session %s: %v", sessionID, err)
			}
		}
	}
}

// computeMetrics calculates all metrics from session events
func (s *Service) computeMetrics(events []*ScoringEvent) *SessionMetrics {
	if len(events) == 0 {
		return &SessionMetrics{}
	}

	// Extract session info from first event
	sessionID := events[0].SessionID
	userID := events[0].UserID

	// Get session metadata (would need to query session entity)
	languageID := "unknown" // TODO: Get from session
	mode := "unknown"       // TODO: Get from session

	// Calculate basic metrics
	duration := s.calculateDuration(events)
	cpm := s.calculateCPM(events, duration)
	twpm := s.calculateTWPM(events, duration)
	kps := s.calculateKPS(events, duration)

	// Calculate accuracy metrics
	rawAccuracy := s.calculateRawAccuracy(events)
	tokenAccuracy := s.calculateTokenAccuracy(events)
	syntaxAccuracy := s.calculateSyntaxAccuracy(events)
	whitespaceAccuracy := s.calculateWhitespaceAccuracy(events)

	// Calculate efficiency metrics
	backspaceRate := s.calculateBackspaceRate(events)
	correctionRate := s.calculateCorrectionRate(events)
	idleTimePercent := s.calculateIdleTimePercent(events)

	// Calculate composite scores
	compositeScore := s.calculateCompositeScore(twpm, rawAccuracy, syntaxAccuracy, backspaceRate, idleTimePercent)
	consistencyScore := s.calculateConsistencyScore(events)
	efficiencyScore := s.calculateEfficiencyScore(events)

	// Analyze patterns
	errorClusters := s.analyzeErrorClusters(events)
	performanceInsights := s.analyzePerformanceInsights(events)
	typingPatterns := s.analyzeTypingPatterns(events)

	// Apply language and mode multipliers
	langMultiplier := s.config.LanguageMultipliers[languageID]
	if langMultiplier == 0 {
		langMultiplier = 1.0
	}
	modeMultiplier := s.config.ModeMultipliers[mode]
	if modeMultiplier == 0 {
		modeMultiplier = 1.0
	}

	compositeScore = int(float64(compositeScore) * langMultiplier * modeMultiplier)

	return &SessionMetrics{
		SessionID:           sessionID,
		UserID:              userID,
		LanguageID:          languageID,
		Mode:                mode,
		DurationMs:          duration,

		CPM:                 cpm,
		TWPM:                twpm,
		KPS:                 kps,

		RawAccuracy:         rawAccuracy,
		TokenAccuracy:       tokenAccuracy,
		SyntaxAccuracy:      syntaxAccuracy,
		WhitespaceAccuracy:  whitespaceAccuracy,

		BackspaceRate:       backspaceRate,
		CorrectionRate:      correctionRate,
		IdleTimePercent:     idleTimePercent,

		CompositeScore:      compositeScore,
		ConsistencyScore:    consistencyScore,
		EfficiencyScore:     efficiencyScore,

		ErrorClusters:       errorClusters,
		PerformanceInsights: performanceInsights,
		TypingPatterns:      typingPatterns,

		SuspiciousActivity:  false, // Will be set by anti-cheat analysis
		CheatFlags:          []string{},
		ConfidenceScore:     1.0, // Will be set by anti-cheat analysis

		CalculatedAt:        time.Now(),
	}
}

// Helper methods for metric calculations
func (s *Service) calculateDuration(events []*ScoringEvent) int64 {
	if len(events) < 2 {
		return 0
	}

	firstTime := events[0].Timestamp
	lastTime := events[len(events)-1].Timestamp

	return lastTime - firstTime
}

func (s *Service) calculateCPM(events []*ScoringEvent, durationMs int64) float64 {
	if durationMs == 0 {
		return 0
	}

	totalChars := 0
	for _, event := range events {
		if event.EventType == "keystroke" && event.Action == "down" {
			totalChars++
		}
	}

	durationMinutes := float64(durationMs) / (1000 * 60)
	return float64(totalChars) / durationMinutes
}

func (s *Service) calculateTWPM(events []*ScoringEvent, durationMs int64) float64 {
	if durationMs == 0 {
		return 0
	}

	totalTokens := 0
	for _, event := range events {
		if event.EventType == "keystroke" && event.Action == "down" && event.ExpectedToken != "" {
			totalTokens++
		}
	}

	durationMinutes := float64(durationMs) / (1000 * 60)
	return float64(totalTokens) / durationMinutes
}

func (s *Service) calculateKPS(events []*ScoringEvent, durationMs int64) float64 {
	if durationMs == 0 {
		return 0
	}

	totalKeystrokes := 0
	for _, event := range events {
		if event.EventType == "keystroke" {
			totalKeystrokes++
		}
	}

	durationSeconds := float64(durationMs) / 1000
	return float64(totalKeystrokes) / durationSeconds
}

func (s *Service) calculateRawAccuracy(events []*ScoringEvent) float64 {
	correct := 0
	total := 0

	for _, event := range events {
		if event.EventType == "keystroke" && event.Action == "down" {
			total++
			if !event.ErrorFlag {
				correct++
			}
		}
	}

	if total == 0 {
		return 0
	}

	return float64(correct) / float64(total)
}

func (s *Service) calculateTokenAccuracy(events []*ScoringEvent) float64 {
	correct := 0
	total := 0

	for _, event := range events {
		if event.EventType == "keystroke" && event.Action == "down" && event.ExpectedToken != "" {
			total++
			if !event.ErrorFlag {
				correct++
			}
		}
	}

	if total == 0 {
		return 0
	}

	return float64(correct) / float64(total)
}

func (s *Service) calculateSyntaxAccuracy(events []*ScoringEvent) float64 {
	// TODO: Implement syntax accuracy calculation based on structural correctness
	return s.calculateRawAccuracy(events) // Placeholder
}

func (s *Service) calculateWhitespaceAccuracy(events []*ScoringEvent) float64 {
	// TODO: Implement whitespace accuracy calculation
	return s.calculateRawAccuracy(events) // Placeholder
}

func (s *Service) calculateBackspaceRate(events []*ScoringEvent) float64 {
	backspaces := 0
	total := 0

	for _, event := range events {
		if event.EventType == "keystroke" {
			total++
			if event.KeyPressed == "Backspace" {
				backspaces++
			}
		}
	}

	if total == 0 {
		return 0
	}

	return float64(backspaces) / float64(total)
}

func (s *Service) calculateCorrectionRate(events []*ScoringEvent) float64 {
	// TODO: Implement correction rate calculation
	return 0.0 // Placeholder
}

func (s *Service) calculateIdleTimePercent(events []*ScoringEvent) float64 {
	// TODO: Implement idle time calculation based on timing gaps
	return 0.0 // Placeholder
}

func (s *Service) calculateCompositeScore(twpm, rawAccuracy, syntaxAccuracy, backspaceRate, idleTimePercent float64) int {
	score := s.config.TWPMWeight * twpm
	score += s.config.RawAccuracyWeight * rawAccuracy * 100
	score += s.config.SyntaxAccuracyWeight * syntaxAccuracy * 100
	score -= s.config.BackspacePenalty * backspaceRate * 100
	score -= s.config.IdleTimePenalty * idleTimePercent * 100

	if score < 0 {
		score = 0
	}

	return int(score)
}

func (s *Service) calculateConsistencyScore(events []*ScoringEvent) float64 {
	// TODO: Implement consistency calculation based on typing rhythm
	return 1.0 // Placeholder
}

func (s *Service) calculateEfficiencyScore(events []*ScoringEvent) float64 {
	// TODO: Implement efficiency calculation
	return 1.0 // Placeholder
}

func (s *Service) analyzeErrorClusters(events []*ScoringEvent) map[string]interface{} {
	// TODO: Implement error clustering analysis
	return make(map[string]interface{})
}

func (s *Service) analyzePerformanceInsights(events []*ScoringEvent) map[string]interface{} {
	// TODO: Implement performance insights generation
	return make(map[string]interface{})
}

func (s *Service) analyzeTypingPatterns(events []*ScoringEvent) map[string]interface{} {
	// TODO: Implement typing pattern analysis
	return make(map[string]interface{})
}

// Database operations
func (s *Service) getSessionEvents(ctx context.Context, sessionID string) ([]*ScoringEvent, error) {
	// TODO: Query SessionEvent entities from Datastore
	return []*ScoringEvent{}, nil
}

func (s *Service) storeMetrics(ctx context.Context, metrics *SessionMetrics) error {
	// TODO: Store ScoringMetrics in Datastore
	return nil
}

func (s *Service) getStoredMetrics(ctx context.Context, sessionID string) (*SessionMetrics, error) {
	// TODO: Retrieve ScoringMetrics from Datastore
	return nil, nil
}

func (s *Service) storeAntiCheatReport(ctx context.Context, report *AntiCheatReport) error {
	// TODO: Store AntiCheatReport in Datastore
	return nil
}

// Anti-cheat analysis
func (s *Service) analyzeAntiCheat(events []*ScoringEvent) *AntiCheatReport {
	report := &AntiCheatReport{
		SessionID: events[0].SessionID,
		UserID:    events[0].UserID,
	}

	// Check for unrealistic KPS
	kps := s.calculateKPS(events, s.calculateDuration(events))
	if kps > s.config.MaxKPS {
		report.UnrealisticKPS = true
		report.SuspiciousEvents = append(report.SuspiciousEvents, "unrealistic_kps")
	}

	// Check for paste events
	for _, event := range events {
		if s.detectPasteEvent(event) {
			report.PasteDetected = true
			report.SuspiciousEvents = append(report.SuspiciousEvents, "paste_detected")
			break
		}
	}

	// Check for auto-type patterns
	if s.detectAutoTypePattern(events) {
		report.AutoTypePattern = true
		report.SuspiciousEvents = append(report.SuspiciousEvents, "auto_type_pattern")
	}

	// Check for anomalous timing
	if s.detectAnomalousTiming(events) {
		report.AnomalousTiming = true
		report.SuspiciousEvents = append(report.SuspiciousEvents, "anomalous_timing")
	}

	// Calculate overall risk level
	report.RiskLevel = s.calculateRiskLevel(report.SuspiciousEvents)
	report.ConfidenceScore = s.calculateConfidenceScore(events)
	report.SuspiciousActivity = len(report.SuspiciousEvents) > 0

	return report
}

func (s *Service) detectPasteEvent(event *ScoringEvent) bool {
	// TODO: Implement paste event detection logic
	return false
}

func (s *Service) detectAutoTypePattern(events []*ScoringEvent) bool {
	// TODO: Implement auto-type pattern detection
	return false
}

func (s *Service) detectAnomalousTiming(events []*ScoringEvent) bool {
	// TODO: Implement anomalous timing detection
	return false
}

func (s *Service) calculateRiskLevel(events []string) string {
	count := len(events)
	switch {
	case count >= 3:
		return "critical"
	case count >= 2:
		return "high"
	case count >= 1:
		return "medium"
	default:
		return "low"
	}
}

func (s *Service) calculateConfidenceScore(events []*ScoringEvent) float64 {
	// TODO: Implement confidence score calculation
	return 1.0
}

func (s *Service) isSessionComplete(events []*ScoringEvent) bool {
	// Check if there's a session_end event
	for _, event := range events {
		if event.EventType == "session_end" {
			return true
		}
	}
	return false
}

// Leaderboard operations

func (s *Service) queryLeaderboard(ctx context.Context, languageID, mode, scope, timeWindow string, limit int) ([]LeaderboardEntry, error) {
	// TODO: Implement leaderboard query logic
	return []LeaderboardEntry{}, nil
}

func (s *Service) updateLeaderboardEntry(ctx context.Context, metrics *SessionMetrics, scope, timeWindow string) error {
	// TODO: Implement leaderboard entry update logic
	return nil
}

func (s *Service) getDefaultConfig() *ScoringConfig {
	return &ScoringConfig{
		Version:             "1.0",
		TWPMWeight:          0.6,
		RawAccuracyWeight:   0.3,
		SyntaxAccuracyWeight: 0.1,
		BackspacePenalty:    0.2,
		IdleTimePenalty:     0.1,
		ConsistencyBonus:    0.05,
		MaxKPS:              10.0,
		MinBurstConsistency: 0.8,
		MaxErrorRate:        0.15,
		MinConfidenceScore:  0.7,
		LanguageMultipliers: map[string]float64{
			"python":     1.0,
			"javascript": 1.1,
			"cpp":        1.2,
			"rust":       1.15,
			"yaml":       0.9,
		},
		ModeMultipliers: map[string]float64{
			"timed":   1.0,
			"zen":     0.8,
			"tutorial": 0.9,
			"accuracy": 1.1,
		},
	}
}
