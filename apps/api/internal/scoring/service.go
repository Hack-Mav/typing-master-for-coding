package scoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/typing-master-for-coding-backend/internal/database"
	"google.golang.org/api/iterator"

	"github.com/typing-master-for-coding-backend/internal/models"
)

// Service handles scoring computation and metrics processing
type Service struct {
	dsClient    DatastoreClient
	cache       *sync.Map
	config      *ScoringConfig
	eventQueue  chan *ScoringEvent
	workerCount int
	processSem  chan struct{}
}

// defaultMaxConcurrentSessions limits the number of sessions scored at the same time.
const defaultMaxConcurrentSessions = 100

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
	SessionID  string `json:"session_id"`
	UserID     string `json:"user_id"`
	LanguageID string `json:"language_id"`
	Mode       string `json:"mode"`
	DurationMs int64  `json:"duration_ms"`

	// Speed metrics
	CPM  float64 `json:"cpm"`
	TWPM float64 `json:"twpm"`
	KPS  float64 `json:"kps"`

	// Accuracy metrics
	RawAccuracy        float64 `json:"raw_accuracy"`
	TokenAccuracy      float64 `json:"token_accuracy"`
	SyntaxAccuracy     float64 `json:"syntax_accuracy"`
	WhitespaceAccuracy float64 `json:"whitespace_accuracy"`

	// Efficiency metrics
	BackspaceRate   float64 `json:"backspace_rate"`
	CorrectionRate  float64 `json:"correction_rate"`
	IdleTimePercent float64 `json:"idle_time_percent"`

	// Composite scores
	CompositeScore   int     `json:"composite_score"`
	ConsistencyScore float64 `json:"consistency_score"`
	EfficiencyScore  float64 `json:"efficiency_score"`

	// Detailed analysis
	ErrorClusters       map[string]interface{} `json:"error_clusters"`
	PerformanceInsights map[string]interface{} `json:"performance_insights"`
	TypingPatterns      map[string]interface{} `json:"typing_patterns"`

	// Anti-cheat analysis
	SuspiciousActivity bool     `json:"suspicious_activity"`
	CheatFlags         []string `json:"cheat_flags"`
	ConfidenceScore    float64  `json:"confidence_score"`

	CalculatedAt time.Time `json:"calculated_at"`
}

// ScoringConfig holds configurable scoring parameters
type ScoringConfig struct {
	Version              string             `json:"version"`
	TWPMWeight           float64            `json:"twpm_weight"`
	RawAccuracyWeight    float64            `json:"raw_accuracy_weight"`
	SyntaxAccuracyWeight float64            `json:"syntax_accuracy_weight"`
	BackspacePenalty     float64            `json:"backspace_penalty"`
	IdleTimePenalty      float64            `json:"idle_time_penalty"`
	ConsistencyBonus     float64            `json:"consistency_bonus"`
	MaxKPS               float64            `json:"max_kps"`
	MinBurstConsistency  float64            `json:"min_burst_consistency"`
	MaxErrorRate         float64            `json:"max_error_rate"`
	MinConfidenceScore   float64            `json:"min_confidence_score"`
	LanguageMultipliers  map[string]float64 `json:"language_multipliers"`
	ModeMultipliers      map[string]float64 `json:"mode_multipliers"`
}

// DatastoreClient defines the interface for the Datastore client operations used in this package.
type DatastoreClient interface {
	Put(ctx context.Context, key *database.Key, src interface{}) (*database.Key, error)
	Get(ctx context.Context, key *database.Key, dst interface{}) error
	Run(ctx context.Context, q *database.Query) database.Iterator
	GetAll(ctx context.Context, q *database.Query, dst interface{}) ([]*database.Key, error)
}

// Service provides scoring and anti-cheat functionalities.

// NewService creates a new scoring service.
func NewService(dsClient DatastoreClient) *Service {
	service := &Service{
		dsClient:    dsClient,
		cache:       &sync.Map{},
		eventQueue:  make(chan *ScoringEvent, 10000), // Buffer for 10k events
		workerCount: 4,
		processSem:  make(chan struct{}, defaultMaxConcurrentSessions),
	}

	// Load default configuration
	service.config = service.getDefaultConfig()

	// Start event processing workers
	service.startWorkers()

	return service
}

// ProcessEvent processes a single scoring event. If the queue is full it blocks
// until the event is accepted or the provided context is cancelled.
// Returns an error if the context is cancelled before the event can be queued.
func (s *Service) ProcessEvent(ctx context.Context, event *ScoringEvent) error {
	select {
	case s.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("event queue full, context cancelled: %w", ctx.Err())
	}
}

// ProcessSession processes all events for a session and computes final metrics.
// It limits the number of concurrent scoring sessions to prevent resource exhaustion.
func (s *Service) ProcessSession(ctx context.Context, sessionID string) (*SessionMetrics, error) {
	select {
	case s.processSem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.processSem }()

	// Get all events for the session
	events, err := s.getSessionEvents(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for session %s", sessionID)
	}

	// Load session metadata
	session, err := s.getSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session metadata: %w", err)
	}

	// Ensure events have user info
	for _, event := range events {
		if event.UserID == "" {
			event.UserID = session.UserID
		}
	}

	// Compute metrics from events
	metrics := s.computeMetrics(session, events)

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

	// Process each session through the bounded scoring pipeline.
	// Process sequentially to avoid unbounded goroutine spawning.
	for sessionID, sessionEventList := range sessionEvents {
		if !s.isSessionComplete(sessionEventList) {
			continue
		}

		processingCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		
		metrics, err := s.ProcessSession(processingCtx, sessionID)
		cancel()
		
		if err != nil {
			log.Printf("Failed to process session %s: %v", sessionID, err)
			continue
		}

		if err := s.UpdateLeaderboard(processingCtx, sessionID); err != nil {
			log.Printf("Failed to update leaderboards for %s: %v", sessionID, err)
		}

		log.Printf("Session %s processed: Score=%d, CPM=%.2f, TWPM=%.2f",
			sessionID, metrics.CompositeScore, metrics.CPM, metrics.TWPM)
	}
}

// computeMetrics calculates all metrics from session events
func (s *Service) computeMetrics(session *models.Session, events []*ScoringEvent) *SessionMetrics {
	if len(events) == 0 {
		return &SessionMetrics{}
	}

	// Extract session info from the session entity and first event
	sessionID := events[0].SessionID
	userID := session.UserID
	if userID == "" {
		userID = events[0].UserID
	}
	languageID := session.LanguageID
	if languageID == "" {
		languageID = "unknown"
	}
	mode := session.Mode
	if mode == "" {
		mode = "unknown"
	}

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
		SessionID:  sessionID,
		UserID:     userID,
		LanguageID: languageID,
		Mode:       mode,
		DurationMs: duration,

		CPM:  cpm,
		TWPM: twpm,
		KPS:  kps,

		RawAccuracy:        rawAccuracy,
		TokenAccuracy:      tokenAccuracy,
		SyntaxAccuracy:     syntaxAccuracy,
		WhitespaceAccuracy: whitespaceAccuracy,

		BackspaceRate:   backspaceRate,
		CorrectionRate:  correctionRate,
		IdleTimePercent: idleTimePercent,

		CompositeScore:   compositeScore,
		ConsistencyScore: consistencyScore,
		EfficiencyScore:  efficiencyScore,

		ErrorClusters:       errorClusters,
		PerformanceInsights: performanceInsights,
		TypingPatterns:      typingPatterns,

		SuspiciousActivity: false, // Will be set by anti-cheat analysis
		CheatFlags:         []string{},
		ConfidenceScore:    1.0, // Will be set by anti-cheat analysis

		CalculatedAt: time.Now(),
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
	correct := 0
	total := 0

	for _, event := range events {
		if event.EventType != "keystroke" || event.Action != "down" {
			continue
		}

		if event.ExpectedToken != "" {
			if isWhitespaceEvent(event.ExpectedToken, event.KeyPressed) {
				continue
			}
			total++
			if !event.ErrorFlag {
				correct++
			}
			continue
		}

		if len(event.KeyPressed) == 1 && !isWhitespaceKey(event.KeyPressed) {
			total++
			if !event.ErrorFlag {
				correct++
			}
		}
	}

	if total == 0 {
		return 1.0
	}

	return float64(correct) / float64(total)
}

func (s *Service) calculateWhitespaceAccuracy(events []*ScoringEvent) float64 {
	correct := 0
	total := 0

	for _, event := range events {
		if event.EventType != "keystroke" || event.Action != "down" {
			continue
		}

		if !isWhitespaceEvent(event.ExpectedToken, event.KeyPressed) {
			continue
		}

		total++
		if !event.ErrorFlag {
			correct++
		}
	}

	if total == 0 {
		return 1.0
	}

	return float64(correct) / float64(total)
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
	backspaces := 0
	errors := 0

	for _, event := range events {
		if event.EventType != "keystroke" {
			continue
		}
		if event.Action == "down" && event.KeyPressed == "Backspace" {
			backspaces++
			continue
		}
		if event.Action == "down" && event.ErrorFlag {
			errors++
		}
	}

	if errors == 0 {
		return 0.0
	}
	if backspaces >= errors {
		return 1.0
	}
	return float64(backspaces) / float64(errors)
}

func (s *Service) calculateIdleTimePercent(events []*ScoringEvent) float64 {
	if len(events) < 2 {
		return 0.0
	}

	totalDuration := events[len(events)-1].Timestamp - events[0].Timestamp
	if totalDuration == 0 {
		return 0.0
	}

	const idleThresholdMs = int64(1000)
	idleTime := int64(0)
	for i := 1; i < len(events); i++ {
		gap := events[i].Timestamp - events[i-1].Timestamp
		if gap > idleThresholdMs {
			idleTime += gap
		}
	}

	if idleTime > totalDuration {
		return 1.0
	}
	return float64(idleTime) / float64(totalDuration)
}

func (s *Service) calculateCompositeScore(twpm, rawAccuracy, syntaxAccuracy, backspaceRate, idleTimePercent float64) int {
	// Normalize TWPM to a 0-100 percentage scale so all components are on the same scale.
	normalizedTWPM := twpm
	if normalizedTWPM > 100.0 {
		normalizedTWPM = 100.0
	}
	if normalizedTWPM < 0.0 {
		normalizedTWPM = 0.0
	}

	score := s.config.TWPMWeight * normalizedTWPM
	score += s.config.RawAccuracyWeight * rawAccuracy * 100
	score += s.config.SyntaxAccuracyWeight * syntaxAccuracy * 100
	score -= s.config.BackspacePenalty * backspaceRate * 100
	score -= s.config.IdleTimePenalty * idleTimePercent * 100

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return int(score)
}

func (s *Service) calculateConsistencyScore(events []*ScoringEvent) float64 {
	if len(events) < 2 {
		return 1.0
	}

	intervals := []float64{}
	for i := 1; i < len(events); i++ {
		if events[i].EventType == "keystroke" && events[i-1].EventType == "keystroke" {
			intervals = append(intervals, float64(events[i].Timestamp-events[i-1].Timestamp))
		}
	}

	if len(intervals) < 2 {
		return 1.0
	}

	sum := 0.0
	for _, interval := range intervals {
		sum += interval
	}
	mean := sum / float64(len(intervals))
	if mean == 0 {
		return 1.0
	}

	variance := 0.0
	for _, interval := range intervals {
		diff := interval - mean
		variance += diff * diff
	}
	variance /= float64(len(intervals))

	stddev := math.Sqrt(variance)
	cv := stddev / mean
	consistency := 1.0 / (1.0 + cv)
	if consistency < 0 {
		return 0
	}
	if consistency > 1 {
		return 1
	}
	return consistency
}

func (s *Service) calculateEfficiencyScore(events []*ScoringEvent) float64 {
	rawAccuracy := s.calculateRawAccuracy(events)
	backspaceRate := s.calculateBackspaceRate(events)
	idleTimePercent := s.calculateIdleTimePercent(events)
	if idleTimePercent > 1.0 {
		idleTimePercent = 1.0
	}

	efficiency := (rawAccuracy + (1.0 - backspaceRate) + (1.0 - idleTimePercent)) / 3.0
	if efficiency < 0 {
		return 0
	}
	if efficiency > 1 {
		return 1
	}
	return efficiency
}

func (s *Service) analyzeErrorClusters(events []*ScoringEvent) map[string]interface{} {
	totalErrors := 0
	errorsByToken := make(map[string]int)
	errorsByKey := make(map[string]int)
	consecutiveClusters := 0
	largestCluster := 0
	currentCluster := 0
	prevError := false
	totalKeystrokes := 0

	for _, event := range events {
		if event.EventType != "keystroke" || event.Action != "down" {
			continue
		}
		totalKeystrokes++
		if event.ErrorFlag {
			totalErrors++
			errorsByToken[event.ExpectedToken]++
			errorsByKey[event.KeyPressed]++
			if !prevError {
				consecutiveClusters++
				currentCluster = 0
			}
			currentCluster++
			if currentCluster > largestCluster {
				largestCluster = currentCluster
			}
			prevError = true
		} else {
			prevError = false
			currentCluster = 0
		}
	}

	errorRate := 0.0
	if totalKeystrokes > 0 {
		errorRate = float64(totalErrors) / float64(totalKeystrokes)
	}

	return map[string]interface{}{
		"total_errors":         totalErrors,
		"error_rate":           errorRate,
		"consecutive_clusters": consecutiveClusters,
		"largest_cluster":      largestCluster,
		"errors_by_token":      errorsByToken,
		"errors_by_key":        errorsByKey,
	}
}

func (s *Service) analyzePerformanceInsights(events []*ScoringEvent) map[string]interface{} {
	if len(events) == 0 {
		return map[string]interface{}{}
	}

	duration := s.calculateDuration(events)
	cpm := s.calculateCPM(events, duration)
	rawAccuracy := s.calculateRawAccuracy(events)
	backspaceRate := s.calculateBackspaceRate(events)
	consistency := s.calculateConsistencyScore(events)

	recommendations := []string{}
	if rawAccuracy < 0.85 {
		recommendations = append(recommendations, "Focus on accuracy before speed")
	}
	if backspaceRate > 0.1 {
		recommendations = append(recommendations, "Reduce backspacing by reading ahead")
	}
	if consistency < 0.5 {
		recommendations = append(recommendations, "Practice a consistent typing rhythm")
	}
	if cpm < 100 {
		recommendations = append(recommendations, "Build speed with easier exercises")
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Great performance — keep challenging yourself")
	}

	return map[string]interface{}{
		"recommendations": recommendations,
		"metrics": map[string]interface{}{
			"cpm":               cpm,
			"raw_accuracy":      rawAccuracy,
			"backspace_rate":    backspaceRate,
			"consistency_score": consistency,
		},
	}
}

func (s *Service) analyzeTypingPatterns(events []*ScoringEvent) map[string]interface{} {
	if len(events) < 2 {
		return map[string]interface{}{}
	}

	intervals := []float64{}
	for i := 1; i < len(events); i++ {
		if events[i].EventType == "keystroke" && events[i-1].EventType == "keystroke" {
			intervals = append(intervals, float64(events[i].Timestamp-events[i-1].Timestamp))
		}
	}

	if len(intervals) == 0 {
		return map[string]interface{}{}
	}

	sum := 0.0
	minInterval := intervals[0]
	maxInterval := intervals[0]
	for _, interval := range intervals {
		sum += interval
		if interval < minInterval {
			minInterval = interval
		}
		if interval > maxInterval {
			maxInterval = interval
		}
	}
	mean := sum / float64(len(intervals))

	variance := 0.0
	for _, interval := range intervals {
		diff := interval - mean
		variance += diff * diff
	}
	variance /= float64(len(intervals))

	totalDuration := events[len(events)-1].Timestamp - events[0].Timestamp
	pauseTime := int64(0)
	for i := 1; i < len(events); i++ {
		gap := events[i].Timestamp - events[i-1].Timestamp
		if gap > 200 {
			pauseTime += gap
		}
	}
	pauseRatio := 0.0
	if totalDuration > 0 {
		pauseRatio = float64(pauseTime) / float64(totalDuration)
	}

	consistency := s.calculateConsistencyScore(events)
	rhythm := "variable"
	if consistency > 0.7 {
		rhythm = "consistent"
	} else if consistency < 0.3 {
		rhythm = "erratic"
	}

	keyFreq := make(map[string]int)
	for _, event := range events {
		if event.EventType == "keystroke" && event.Action == "down" {
			keyFreq[event.KeyPressed]++
		}
	}
	dominantKey := ""
	dominantCount := 0
	for key, count := range keyFreq {
		if count > dominantCount {
			dominantCount = count
			dominantKey = key
		}
	}

	return map[string]interface{}{
		"average_interval_ms": mean,
		"min_interval_ms":     minInterval,
		"max_interval_ms":     maxInterval,
		"interval_variance":   variance,
		"pause_ratio":         pauseRatio,
		"rhythm":              rhythm,
		"consistency_score":   consistency,
		"dominant_key":        dominantKey,
		"dominant_key_count":  dominantCount,
	}
}

func isWhitespaceKey(key string) bool {
	if key == " " || key == "\t" || key == "\n" || key == "\r" ||
		key == "Space" || key == "Enter" || key == "Return" || key == "Tab" {
		return true
	}
	if len(key) == 1 {
		return unicode.IsSpace(rune(key[0]))
	}
	return false
}

func isWhitespaceEvent(expectedToken, key string) bool {
	if expectedToken != "" {
		return strings.TrimSpace(expectedToken) == ""
	}
	return isWhitespaceKey(key)
}

// Database operations
func (s *Service) getSessionEvents(ctx context.Context, sessionID string) ([]*ScoringEvent, error) {
	// Query SessionEvent entities from Datastore
	query := database.NewQuery("SessionEvent").
		FilterField("SessionID", "=", sessionID).
		Order("TimestampMs")

	var events []*ScoringEvent
	iter := s.dsClient.Run(ctx, query)

	for {
		var sessionEvent struct {
			SessionID      string
			TimestampMs    int64
			KeyPressed     string
			Action         string
			CursorPosition int
			ErrorFlag      bool
			ExpectedToken  string
			Metadata       map[string]interface{}
		}

		_, err := iter.Next(&sessionEvent)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate events: %w", err)
		}

		// Convert to ScoringEvent
		event := &ScoringEvent{
			SessionID:     sessionEvent.SessionID,
			EventType:     "keystroke",
			Timestamp:     sessionEvent.TimestampMs,
			KeyPressed:    sessionEvent.KeyPressed,
			Action:        sessionEvent.Action,
			CursorPos:     sessionEvent.CursorPosition,
			ErrorFlag:     sessionEvent.ErrorFlag,
			ExpectedToken: sessionEvent.ExpectedToken,
			Metadata:      sessionEvent.Metadata,
		}
		events = append(events, event)
	}

	return events, nil
}

func (s *Service) getSession(ctx context.Context, sessionID string) (*models.Session, error) {
	key := database.NameKey("Session", sessionID, nil)
	var session models.Session
	if err := s.dsClient.Get(ctx, key, &session); err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

func (s *Service) storeMetrics(ctx context.Context, metrics *SessionMetrics) error {
	// Store ScoringMetrics in Datastore
	key := database.NameKey("ScoringMetrics", metrics.SessionID, nil)

	_, err := s.dsClient.Put(ctx, key, metrics)
	if err != nil {
		return fmt.Errorf("failed to store metrics: %w", err)
	}

	// Also store as Result for backward compatibility
	result := &struct {
		SessionID      string
		CPM            float64
		TWPM           float64
		RawAccuracy    float64
		TokenAccuracy  float64
		SyntaxAccuracy float64
		BackspaceRate  float64
		CompositeScore int
		Breakdown      map[string]interface{}
		CreatedAt      time.Time
	}{
		SessionID:      metrics.SessionID,
		CPM:            metrics.CPM,
		TWPM:           metrics.TWPM,
		RawAccuracy:    metrics.RawAccuracy,
		TokenAccuracy:  metrics.TokenAccuracy,
		SyntaxAccuracy: metrics.SyntaxAccuracy,
		BackspaceRate:  metrics.BackspaceRate,
		CompositeScore: metrics.CompositeScore,
		Breakdown: map[string]interface{}{
			"error_clusters":       metrics.ErrorClusters,
			"performance_insights": metrics.PerformanceInsights,
			"typing_patterns":      metrics.TypingPatterns,
		},
		CreatedAt: metrics.CalculatedAt,
	}

	resultKey := database.NameKey("Result", metrics.SessionID, nil)
	_, err = s.dsClient.Put(ctx, resultKey, result)
	if err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	return nil
}

func (s *Service) getStoredMetrics(ctx context.Context, sessionID string) (*SessionMetrics, error) {
	// Retrieve ScoringMetrics from Datastore
	key := database.NameKey("ScoringMetrics", sessionID, nil)

	var metrics SessionMetrics
	err := s.dsClient.Get(ctx, key, &metrics)
	if err != nil {
		if err == database.ErrNoSuchEntity {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return &metrics, nil
}

func (s *Service) storeAntiCheatReport(ctx context.Context, report *AntiCheatReport) error {
	// Store AntiCheatReport in Datastore
	reportID := fmt.Sprintf("%s_%d", report.SessionID, time.Now().Unix())
	key := database.NameKey("AntiCheatReport", reportID, nil)

	_, err := s.dsClient.Put(ctx, key, report)
	if err != nil {
		return fmt.Errorf("failed to store anti-cheat report: %w", err)
	}

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
	// Calculate time window boundaries
	windowStart, windowEnd := s.calculateTimeWindow(timeWindow)

	// Query leaderboard entities
	query := database.NewQuery("Leaderboard").
		FilterField("LanguageID", "=", languageID).
		FilterField("Mode", "=", mode).
		FilterField("Scope", "=", scope).
		FilterField("TimeWindow", "=", timeWindow).
		FilterField("RecordedAt", ">=", windowStart).
		FilterField("RecordedAt", "<=", windowEnd).
		Order("-Score").
		Limit(limit)

	var leaderboardEntities []struct {
		UserID             string
		Score              int
		CPM                float64
		TWPM               float64
		Accuracy           float64
		Badge              string
		IsVerified         bool
		VerificationMethod string
		RecordedAt         time.Time
	}

	_, err := s.dsClient.GetAll(ctx, query, &leaderboardEntities)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}

	// Convert to LeaderboardEntry
	entries := make([]LeaderboardEntry, len(leaderboardEntities))
	for i, entity := range leaderboardEntities {
		entries[i] = LeaderboardEntry{
			UserID:             entity.UserID,
			Username:           fmt.Sprintf("user_%s", entity.UserID[:8]),
			Score:              entity.Score,
			Rank:               i + 1,
			CPM:                entity.CPM,
			TWPM:               entity.TWPM,
			Accuracy:           entity.Accuracy,
			Badge:              entity.Badge,
			IsVerified:         entity.IsVerified,
			VerificationMethod: entity.VerificationMethod,
			LastActive:         entity.RecordedAt,
		}
	}

	return entries, nil
}

func (s *Service) updateLeaderboardEntry(ctx context.Context, metrics *SessionMetrics, scope, timeWindow string) error {
	// Calculate time window boundaries
	windowStart, windowEnd := s.calculateTimeWindow(timeWindow)

	// Create leaderboard entry
	entryID := fmt.Sprintf("%s_%s_%s_%s_%d",
		metrics.UserID, metrics.LanguageID, metrics.Mode, timeWindow, time.Now().Unix())

	entry := struct {
		UserID             string
		SessionID          string
		LanguageID         string
		Mode               string
		Scope              string
		TimeWindow         string
		Username           string // Added for direct storage
		Handle             string // Added for direct storage
		Rank               int
		Score              int
		CPM                float64
		TWPM               float64
		Accuracy           float64
		MetricsSnapshot    map[string]interface{}
		Badge              string
		IsVerified         bool
		VerificationMethod string
		RecordedAt         time.Time
		WindowStart        time.Time
		WindowEnd          time.Time
	}{
		UserID:     metrics.UserID,
		SessionID:  metrics.SessionID,
		LanguageID: metrics.LanguageID,
		Mode:       metrics.Mode,
		Scope:      scope,
		TimeWindow: timeWindow,
		Username:   "", // Will be populated if available
		Handle:     "", // Will be populated if available
		Rank:       0,  // Will be calculated later
		Score:      metrics.CompositeScore,
		CPM:        metrics.CPM,
		TWPM:       metrics.TWPM,
		Accuracy:   metrics.RawAccuracy,
		MetricsSnapshot: map[string]interface{}{
			"kps":               metrics.KPS,
			"token_accuracy":    metrics.TokenAccuracy,
			"syntax_accuracy":   metrics.SyntaxAccuracy,
			"consistency_score": metrics.ConsistencyScore,
			"efficiency_score":  metrics.EfficiencyScore,
		},
		Badge:              "",
		IsVerified:         metrics.ConfidenceScore >= s.config.MinConfidenceScore,
		VerificationMethod: "auto",
		RecordedAt:         time.Now(),
		WindowStart:        windowStart,
		WindowEnd:          windowEnd,
	}

	key := database.NameKey("Leaderboard", entryID, nil)
	_, err := s.dsClient.Put(ctx, key, &entry)
	if err != nil {
		return fmt.Errorf("failed to update leaderboard entry: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("leaderboard:%s:%s:%s:%s", metrics.LanguageID, metrics.Mode, scope, timeWindow)
	s.cache.Delete(cacheKey)

	return nil
}

// calculateTimeWindow calculates start and end times for a given time window
func (s *Service) calculateTimeWindow(timeWindow string) (time.Time, time.Time) {
	now := time.Now()
	var windowStart time.Time

	switch timeWindow {
	case "daily":
		windowStart = now.AddDate(0, 0, -1)
	case "weekly":
		windowStart = now.AddDate(0, 0, -7)
	case "monthly":
		windowStart = now.AddDate(0, -1, 0)
	case "all_time":
		windowStart = time.Time{} // Beginning of time
	default:
		windowStart = now.AddDate(0, 0, -7) // Default to weekly
	}

	return windowStart, now
}

func (s *Service) getDefaultConfig() *ScoringConfig {
	return &ScoringConfig{
		Version:              "1.0",
		TWPMWeight:           0.6,
		RawAccuracyWeight:    0.3,
		SyntaxAccuracyWeight: 0.1,
		BackspacePenalty:     0.2,
		IdleTimePenalty:      0.1,
		ConsistencyBonus:     0.05,
		MaxKPS:               10.0,
		MinBurstConsistency:  0.8,
		MaxErrorRate:         0.15,
		MinConfidenceScore:   0.7,
		LanguageMultipliers: map[string]float64{
			"python":     1.0,
			"javascript": 1.1,
			"cpp":        1.2,
			"rust":       1.15,
			"yaml":       0.9,
		},
		ModeMultipliers: map[string]float64{
			"timed":    1.0,
			"zen":      0.8,
			"tutorial": 0.9,
			"accuracy": 1.1,
		},
	}
}

// LeaderboardService defines the interface for leaderboard operations.
type LeaderboardService interface {
	GetLeaderboard(ctx context.Context, req *LeaderboardRequest) (*LeaderboardResponse, error)
	UpdateLeaderboard(ctx context.Context, sessionID string) error
	GetUserRank(ctx context.Context, userID, languageID, mode, scope, timeWindow string) (*UserRank, error)
	GetLeaderboardTrends(ctx context.Context, req *LeaderboardTrendsRequest) (*LeaderboardTrendsResponse, error)
}
