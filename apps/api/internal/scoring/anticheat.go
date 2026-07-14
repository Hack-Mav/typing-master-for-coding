package scoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"google.golang.org/api/iterator"
)

// AntiCheatReport represents the results of anti-cheat analysis
type AntiCheatReport struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`

	// Detection results
	PasteDetected   bool `json:"paste_detected"`
	UnrealisticKPS  bool `json:"unrealistic_kps"`
	AutoTypePattern bool `json:"auto_type_pattern"`
	WindowFocusLost bool `json:"window_focus_lost"`
	AnomalousTiming bool `json:"anomalous_timing"`

	// Analysis data
	KPSVariance        float64 `json:"kps_variance"`
	BurstConsistency   float64 `json:"burst_consistency"`
	ErrorPatternScore  float64 `json:"error_pattern_score"`
	StatisticalAnomaly float64 `json:"statistical_anomaly"`

	// Evidence
	SuspiciousEvents   []string               `json:"suspicious_events"`
	EvidenceDetails    map[string]interface{} `json:"evidence_details"`
	ConfidenceScore    float64                `json:"confidence_score"`
	RiskLevel          string                 `json:"risk_level"` // "low", "medium", "high", "critical"
	SuspiciousActivity bool                   `json:"suspicious_activity"`
	CheatFlags         []string               `json:"cheat_flags"`

	// Actions taken
	FlaggedForReview    bool `json:"flagged_for_review"`
	ScoreInvalidated    bool `json:"score_invalidated"`
	LeaderboardExcluded bool `json:"leaderboard_excluded"`

	// Timestamps
	DetectedAt time.Time  `json:"detected_at"`
	ReviewedAt *time.Time `json:"reviewed_at"`
	ReviewedBy string     `json:"reviewed_by"`
}

// AntiCheatConfig holds configuration for anti-cheat detection
type AntiCheatConfig struct {
	Enabled              bool                 `json:"enabled"`
	StrictMode           bool                 `json:"strict_mode"`
	VerificationRequired bool                 `json:"verification_required"`
	VerificationMethods  []string             `json:"verification_methods"`
	Thresholds           *DetectionThresholds `json:"thresholds"`
}

// DetectionThresholds holds threshold values for various cheat detection methods
type DetectionThresholds struct {
	MaxKPS                         float64 `json:"max_kps"`
	MinBurstConsistency            float64 `json:"min_burst_consistency"`
	MaxErrorRate                   float64 `json:"max_error_rate"`
	MinIdleTimePercent             float64 `json:"min_idle_time_percent"`
	MaxPasteEventRate              float64 `json:"max_paste_event_rate"`
	MinConfidenceScore             float64 `json:"min_confidence_score"`
	MaxAnomalousEvents             int     `json:"max_anomalous_events"`
	StatisticalThreshold           float64 `json:"statistical_threshold"`
	ExpectedKPS                    float64 `json:"expected_kps"`
	ExpectedAccuracy               float64 `json:"expected_accuracy"`
	ExpectedErrorRate              float64 `json:"expected_error_rate"`
	AccuracyDeviationThreshold     float64 `json:"accuracy_deviation_threshold"`
	ErrorRateDeviationThreshold    float64 `json:"error_rate_deviation_threshold"`
	ExpectedBurstConsistency       float64 `json:"expected_burst_consistency"`
	StatisticalAnomalyThreshold    float64 `json:"statistical_anomaly_threshold"`
	WindowFocusLostGapMs           int64   `json:"window_focus_lost_gap_ms"`
	WindowFocusLostMinCount        int     `json:"window_focus_lost_min_count"`
	PasteKPSMultiplier             float64 `json:"paste_kps_multiplier"`
	PasteMinSequenceLength         int     `json:"paste_min_sequence_length"`
	PasteIntervalVarianceThreshold float64 `json:"paste_interval_variance_threshold"`
}

// AntiCheatService handles anti-cheat detection and analysis
type AntiCheatService struct {
	dsClient   DatastoreClient
	config     *AntiCheatConfig
	thresholds *DetectionThresholds
}

// NewAntiCheatService creates a new anti-cheat service
func NewAntiCheatService(dsClient DatastoreClient) *AntiCheatService {
	config := getDefaultAntiCheatConfig()
	service := &AntiCheatService{
		dsClient:   dsClient,
		config:     config,
		thresholds: config.Thresholds,
	}

	return service
}

// AnalyzeSession performs comprehensive anti-cheat analysis on a session
func (s *AntiCheatService) AnalyzeSession(ctx context.Context, sessionID string) (*AntiCheatReport, error) {
	// Get session events
	events, err := s.getSessionEvents(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session events: %w", err)
	}

	if len(events) == 0 {
		return &AntiCheatReport{
			SessionID:       sessionID,
			ConfidenceScore: 1.0,
		}, nil
	}

	// Perform all detection methods
	report := &AntiCheatReport{
		SessionID:       sessionID,
		UserID:          events[0].UserID,
		EvidenceDetails: make(map[string]interface{}),
	}

	// 1. Paste event detection
	report.PasteDetected = s.detectPasteEvents(events)
	if report.PasteDetected {
		report.SuspiciousEvents = append(report.SuspiciousEvents, "paste_detected")
		report.EvidenceDetails["paste_events"] = s.getPasteEventEvidence(events)
	}

	// 2. Unrealistic KPS detection
	kps := s.calculateKPS(events)
	report.UnrealisticKPS = len(events) >= 10 && kps > s.thresholds.MaxKPS
	if report.UnrealisticKPS {
		report.SuspiciousEvents = append(report.SuspiciousEvents, "unrealistic_kps")
		report.EvidenceDetails["kps"] = kps
		report.EvidenceDetails["max_kps_threshold"] = s.thresholds.MaxKPS
	}

	// 3. Auto-type pattern detection
	report.AutoTypePattern = s.detectAutoTypePattern(events)
	if report.AutoTypePattern {
		report.SuspiciousEvents = append(report.SuspiciousEvents, "auto_type_pattern")
		report.EvidenceDetails["pattern_score"] = s.calculatePatternScore(events)
	}

	// 4. Window focus tracking (placeholder for future implementation)
	report.WindowFocusLost = s.detectWindowFocusLoss(events)
	if report.WindowFocusLost {
		report.SuspiciousEvents = append(report.SuspiciousEvents, "window_focus_lost")
	}

	// 5. Statistical anomaly detection
	report.AnomalousTiming = s.detectStatisticalAnomalies(events)
	if report.AnomalousTiming {
		report.SuspiciousEvents = append(report.SuspiciousEvents, "anomalous_timing")
		report.EvidenceDetails["statistical_score"] = s.calculateStatisticalScore(events)
	}

	// Calculate overall metrics
	report.KPSVariance = s.calculateKPSVariance(events)
	report.BurstConsistency = s.calculateBurstConsistency(events)
	report.ErrorPatternScore = s.calculateErrorPatternScore(events)
	report.StatisticalAnomaly = s.calculateOverallStatisticalAnomaly(events)

	// Determine risk level and confidence
	report.RiskLevel = s.calculateRiskLevel(len(report.SuspiciousEvents))
	report.ConfidenceScore = s.calculateConfidenceScore(events, report.SuspiciousEvents)

	// Overall suspicious activity flag
	report.SuspiciousActivity = len(report.SuspiciousEvents) > 0 && report.ConfidenceScore >= s.thresholds.MinConfidenceScore

	// Determine actions based on risk level
	s.determineActions(report)

	// Store report for audit
	if err := s.storeReport(ctx, report); err != nil {
		log.Printf("Failed to store anti-cheat report: %v", err)
	}

	return report, nil
}

// DetectPasteEvents identifies potential paste events in typing sessions
func (s *AntiCheatService) detectPasteEvents(events []*ScoringEvent) bool {
	minSeqLen := s.thresholds.PasteMinSequenceLength
	if len(events) < minSeqLen {
		return false
	}

	pasteKPS := s.thresholds.MaxKPS * s.thresholds.PasteKPSMultiplier
	if pasteKPS <= 0 {
		return false
	}

	// Look for a contiguous sequence of keystrokes that are both extremely fast
	// and have machine-like consistency (near-zero timing variance). This avoids
	// flagging skilled typists who may produce short, fast bursts with natural jitter.
	for i := 0; i <= len(events)-minSeqLen; i++ {
		window := events[i : i+minSeqLen]
		if s.calculateBurstKPS(window) < pasteKPS {
			continue
		}
		if s.calculateIntervalVariance(window) <= s.thresholds.PasteIntervalVarianceThreshold {
			return true
		}
	}

	return false
}

// DetectAutoTypePattern identifies bot-like typing patterns
func (s *AntiCheatService) detectAutoTypePattern(events []*ScoringEvent) bool {
	if len(events) < 50 {
		return false
	}

	kps := s.calculateKPS(events)
	autoTypeKPS := s.thresholds.MaxKPS * 0.75

	// Check for perfect rhythm consistency (bots often have exact timing)
	intervalVariance := s.calculateIntervalVariance(events)

	// Too perfect rhythm at high speed is suspicious
	if intervalVariance < 0.1 && kps > autoTypeKPS { // Less than 10% variance
		return true
	}

	// Check for lack of natural typing pauses at high speed
	pauseRatio := s.calculatePauseRatio(events)
	if pauseRatio < 0.05 && kps > autoTypeKPS { // Less than 5% pauses
		return true
	}

	return false
}

// DetectWindowFocusLoss detects if user switched away from typing window
func (s *AntiCheatService) detectWindowFocusLoss(events []*ScoringEvent) bool {
	// Prefer explicit client-side focus/blur metadata when available.
	for _, event := range events {
		if event.Metadata == nil {
			continue
		}
		if focused, ok := event.Metadata["window_focused"].(bool); ok && !focused {
			return true
		}
		if blurred, ok := event.Metadata["window_blurred"].(bool); ok && blurred {
			return true
		}
	}

	// Fall back to detecting multiple long gaps in typing.
	longGaps := 0
	gapMs := s.thresholds.WindowFocusLostGapMs
	for i := 1; i < len(events); i++ {
		gap := events[i].Timestamp - events[i-1].Timestamp
		if gap > gapMs {
			longGaps++
		}
	}

	return longGaps >= s.thresholds.WindowFocusLostMinCount
}

// DetectStatisticalAnomalies uses statistical analysis to find outliers
func (s *AntiCheatService) detectStatisticalAnomalies(events []*ScoringEvent) bool {
	if len(events) < 20 {
		return false
	}

	// Calculate various statistical measures
	kps := s.calculateKPS(events)
	accuracy := s.calculateAccuracy(events)
	errorRate := s.calculateErrorRate(events)
	burstConsistency := s.calculateBurstConsistency(events)

	deviations := 0

	if math.Abs(kps-s.thresholds.ExpectedKPS) > s.thresholds.StatisticalThreshold {
		deviations++
	}

	if math.Abs(accuracy-s.thresholds.ExpectedAccuracy) > s.thresholds.AccuracyDeviationThreshold {
		deviations++
	}

	if math.Abs(errorRate-s.thresholds.ExpectedErrorRate) > s.thresholds.ErrorRateDeviationThreshold {
		deviations++
	}

	if burstConsistency < s.thresholds.ExpectedBurstConsistency {
		deviations++
	}

	if kps > s.thresholds.MaxKPS {
		deviations++
	}

	return deviations >= int(s.thresholds.StatisticalAnomalyThreshold)
}

// Helper calculation methods

func (s *AntiCheatService) calculateKPS(events []*ScoringEvent) float64 {
	if len(events) == 0 {
		return 0
	}

	// Get total duration in seconds
	firstTime := events[0].Timestamp
	lastTime := events[len(events)-1].Timestamp
	durationSeconds := float64(lastTime-firstTime) / 1000

	if durationSeconds == 0 {
		return 0
	}

	// Count keystroke events
	keystrokeCount := 0
	for _, event := range events {
		if event.EventType == "keystroke" {
			keystrokeCount++
		}
	}

	return float64(keystrokeCount) / durationSeconds
}

func (s *AntiCheatService) calculateAccuracy(events []*ScoringEvent) float64 {
	correct := 0
	total := 0

	for _, event := range events {
		if event.EventType == "keystroke" && (event.Action == "down" || event.Action == "") {
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

func (s *AntiCheatService) calculateErrorRate(events []*ScoringEvent) float64 {
	errors := 0
	total := 0

	for _, event := range events {
		if event.EventType == "keystroke" && (event.Action == "down" || event.Action == "") {
			total++
			if event.ErrorFlag {
				errors++
			}
		}
	}

	if total == 0 {
		return 0
	}

	return float64(errors) / float64(total)
}

func (s *AntiCheatService) calculateBurstKPS(events []*ScoringEvent) float64 {
	if len(events) < 2 {
		return 0
	}

	firstTime := events[0].Timestamp
	lastTime := events[len(events)-1].Timestamp
	durationSeconds := float64(lastTime-firstTime) / 1000

	if durationSeconds == 0 {
		return 0
	}

	keystrokeCount := 0
	for _, event := range events {
		if event.EventType == "keystroke" {
			keystrokeCount++
		}
	}

	return float64(keystrokeCount) / durationSeconds
}

func (s *AntiCheatService) calculateIntervalVariance(events []*ScoringEvent) float64 {
	if len(events) < 3 {
		return 0
	}

	// Calculate intervals between keystrokes
	intervals := []float64{}
	for i := 1; i < len(events); i++ {
		if events[i].EventType == "keystroke" && events[i-1].EventType == "keystroke" {
			interval := float64(events[i].Timestamp - events[i-1].Timestamp)
			intervals = append(intervals, interval)
		}
	}

	if len(intervals) < 2 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, interval := range intervals {
		sum += interval
	}
	mean := sum / float64(len(intervals))

	// Calculate variance
	variance := 0.0
	for _, interval := range intervals {
		variance += (interval - mean) * (interval - mean)
	}
	variance /= float64(len(intervals))

	if mean == 0 {
		return 0
	}

	// Return coefficient of variation (std dev / mean)
	return math.Sqrt(variance) / mean
}

func (s *AntiCheatService) calculatePauseRatio(events []*ScoringEvent) float64 {
	if len(events) < 2 {
		return 0
	}

	totalTime := float64(events[len(events)-1].Timestamp - events[0].Timestamp)
	pauseTime := 0.0

	for i := 1; i < len(events); i++ {
		gap := events[i].Timestamp - events[i-1].Timestamp
		if gap > 200 { // Pauses longer than 200ms
			pauseTime += float64(gap)
		}
	}

	return pauseTime / totalTime
}

func (s *AntiCheatService) calculateBurstConsistency(events []*ScoringEvent) float64 {
	// Calculate consistency of typing bursts
	// TODO: Implement burst consistency calculation
	return 0.85 // Placeholder
}

func (s *AntiCheatService) calculateErrorPatternScore(events []*ScoringEvent) float64 {
	// Analyze error patterns for suspicious clustering
	// TODO: Implement error pattern analysis
	return 0.9 // Placeholder
}

func (s *AntiCheatService) calculateStatisticalScore(events []*ScoringEvent) float64 {
	// Calculate overall statistical anomaly score
	kps := s.calculateKPS(events)
	accuracy := s.calculateAccuracy(events)

	// Z-score like calculation against configured expected values
	kpsDeviation := math.Abs(kps-s.thresholds.ExpectedKPS) / s.thresholds.StatisticalThreshold
	accuracyDeviation := math.Abs(accuracy-s.thresholds.ExpectedAccuracy) / s.thresholds.AccuracyDeviationThreshold

	return kpsDeviation + accuracyDeviation
}

func (s *AntiCheatService) calculateOverallStatisticalAnomaly(events []*ScoringEvent) float64 {
	score := s.calculateStatisticalScore(events)

	// Check for multiple anomalies
	anomalies := 0
	if score > s.thresholds.StatisticalAnomalyThreshold {
		anomalies++
	}

	// Additional checks
	if s.detectAutoTypePattern(events) {
		anomalies++
	}

	if s.detectPasteEvents(events) {
		anomalies++
	}

	return float64(anomalies)
}

func (s *AntiCheatService) calculateKPSVariance(events []*ScoringEvent) float64 {
	// Calculate variance in keystrokes per second over time windows
	// TODO: Implement KPS variance calculation
	return 0.5 // Placeholder
}

func (s *AntiCheatService) calculateRiskLevel(suspiciousEventCount int) string {
	switch {
	case suspiciousEventCount >= 3:
		return "critical"
	case suspiciousEventCount >= 2:
		return "high"
	case suspiciousEventCount >= 1:
		return "medium"
	default:
		return "low"
	}
}

func (s *AntiCheatService) calculateConfidenceScore(events []*ScoringEvent, suspiciousEvents []string) float64 {
	baseConfidence := 0.5

	// Increase confidence based on number of suspicious events
	eventCount := len(suspiciousEvents)
	baseConfidence += float64(eventCount) * 0.2

	// Increase confidence based on session length (longer sessions = more confident)
	sessionLength := len(events)
	if sessionLength > 100 {
		baseConfidence += 0.2
	} else if sessionLength > 50 {
		baseConfidence += 0.1
	}

	// Cap at 1.0
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

func (s *AntiCheatService) determineActions(report *AntiCheatReport) {
	switch report.RiskLevel {
	case "critical":
		report.FlaggedForReview = true
		report.ScoreInvalidated = true
		report.LeaderboardExcluded = true
	case "high":
		report.FlaggedForReview = true
		report.ScoreInvalidated = report.ConfidenceScore > 0.8
		report.LeaderboardExcluded = report.ConfidenceScore > 0.7
	case "medium":
		report.FlaggedForReview = report.ConfidenceScore > 0.6
		report.ScoreInvalidated = false
		report.LeaderboardExcluded = false
	case "low":
		// No actions needed
	}
}

func (s *AntiCheatService) getPasteEventEvidence(events []*ScoringEvent) interface{} {
	// Return evidence of paste events
	pasteSequences := 0
	minSeqLen := s.thresholds.PasteMinSequenceLength
	pasteKPS := s.thresholds.MaxKPS * s.thresholds.PasteKPSMultiplier
	for i := 0; i <= len(events)-minSeqLen; i++ {
		window := events[i : i+minSeqLen]
		if s.calculateBurstKPS(window) < pasteKPS {
			continue
		}
		if s.calculateIntervalVariance(window) <= s.thresholds.PasteIntervalVarianceThreshold {
			pasteSequences++
		}
	}
	return map[string]interface{}{
		"paste_sequences": pasteSequences,
		"threshold":       pasteKPS,
	}
}

func (s *AntiCheatService) calculatePatternScore(events []*ScoringEvent) float64 {
	// Calculate how "perfect" the typing pattern is (lower is more suspicious)
	variance := s.calculateIntervalVariance(events)
	pauseRatio := s.calculatePauseRatio(events)

	// Perfect patterns have low variance and low pause ratio
	patternScore := variance + pauseRatio

	// Lower score = more suspicious
	if patternScore < 0.1 {
		return 0.2 // Very suspicious
	} else if patternScore < 0.3 {
		return 0.5 // Moderately suspicious
	}

	return 0.9 // Normal pattern
}

// Database operations
func (s *AntiCheatService) getSessionEvents(ctx context.Context, sessionID string) ([]*ScoringEvent, error) {
	query := database.NewQuery("SessionEvent").
		FilterField("SessionID", "=", sessionID).
		Order("TimestampMs")

	var events []*ScoringEvent
	iter := s.dsClient.Run(ctx, query)

	for {
		var event ScoringEvent
		_, err := iter.Next(&event)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate events: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

func (s *AntiCheatService) storeReport(ctx context.Context, report *AntiCheatReport) error {
	// Set timestamps
	report.DetectedAt = time.Now()

	// TODO: Store AntiCheatReport in Datastore
	key := database.NameKey("AntiCheatReport", fmt.Sprintf("%s_%d", report.SessionID, time.Now().Unix()), nil)

	_, err := s.dsClient.Put(ctx, key, report)
	return err
}

// Tournament verification methods

// VerifyTournamentParticipant performs verification for tournament participants
func (s *AntiCheatService) VerifyTournamentParticipant(ctx context.Context, tournamentID, userID string, verificationMethod string) (*VerificationResult, error) {
	result := &VerificationResult{
		TournamentID: tournamentID,
		UserID:       userID,
		Method:       verificationMethod,
		VerifiedAt:   time.Now(),
	}

	switch verificationMethod {
	case "webcam":
		result.Status = s.verifyWebcam(ctx, userID)
	case "hid":
		result.Status = s.verifyHID(ctx, userID)
	case "both":
		webcamStatus := s.verifyWebcam(ctx, userID)
		hidStatus := s.verifyHID(ctx, userID)
		if webcamStatus == "verified" && hidStatus == "verified" {
			result.Status = "verified"
		} else {
			result.Status = "failed"
		}
	default:
		result.Status = "unsupported_method"
	}

	// Store verification result
	err := s.storeVerificationResult(ctx, result)
	if err != nil {
		log.Printf("Failed to store verification result: %v", err)
	}

	return result, nil
}

func (s *AntiCheatService) verifyWebcam(ctx context.Context, userID string) string {
	// TODO: Implement webcam verification logic
	// This would involve WebRTC camera access verification
	return "verified" // Placeholder
}

func (s *AntiCheatService) verifyHID(ctx context.Context, userID string) string {
	// TODO: Implement HID verification logic
	// This would involve checking for human interface device activity
	return "verified" // Placeholder
}

// VerificationResult represents the result of participant verification
type VerificationResult struct {
	TournamentID string                 `json:"tournament_id"`
	UserID       string                 `json:"user_id"`
	Method       string                 `json:"method"`
	Status       string                 `json:"status"` // "verified", "failed", "unsupported_method"
	Details      map[string]interface{} `json:"details"`
	VerifiedAt   time.Time              `json:"verified_at"`
}

func (s *AntiCheatService) storeVerificationResult(ctx context.Context, result *VerificationResult) error {
	// TODO: Store verification result (could be part of TournamentParticipant entity)
	return nil
}

// Helper functions

func getDefaultAntiCheatConfig() *AntiCheatConfig {
	return &AntiCheatConfig{
		Enabled:              true,
		StrictMode:           false,
		VerificationRequired: false,
		VerificationMethods:  []string{"webcam", "hid"},
		Thresholds: &DetectionThresholds{
			MaxKPS:                         12.0,
			MinBurstConsistency:            0.7,
			MaxErrorRate:                   0.15,
			MinIdleTimePercent:             0.05,
			MaxPasteEventRate:              0.02,
			MinConfidenceScore:             0.6,
			MaxAnomalousEvents:             3,
			StatisticalThreshold:           2.5,
			ExpectedKPS:                    6.0,
			ExpectedAccuracy:               0.95,
			ExpectedErrorRate:              0.05,
			AccuracyDeviationThreshold:     0.1,
			ErrorRateDeviationThreshold:    0.05,
			ExpectedBurstConsistency:       0.8,
			StatisticalAnomalyThreshold:    2.0,
			WindowFocusLostGapMs:           5000,
			WindowFocusLostMinCount:        3,
			PasteKPSMultiplier:             3.0,
			PasteMinSequenceLength:         5,
			PasteIntervalVarianceThreshold: 0.1,
		},
	}
}
