package scoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"cloud.google.com/go/datastore"
)

// AntiCheatReport represents the results of anti-cheat analysis
type AntiCheatReport struct {
	SessionID           string                 `json:"session_id"`
	UserID              string                 `json:"user_id"`

	// Detection results
	PasteDetected       bool                   `json:"paste_detected"`
	UnrealisticKPS      bool                   `json:"unrealistic_kps"`
	AutoTypePattern     bool                   `json:"auto_type_pattern"`
	WindowFocusLost     bool                   `json:"window_focus_lost"`
	AnomalousTiming     bool                   `json:"anomalous_timing"`

	// Analysis data
	KPSVariance         float64                `json:"kps_variance"`
	BurstConsistency    float64                `json:"burst_consistency"`
	ErrorPatternScore   float64                `json:"error_pattern_score"`
	StatisticalAnomaly  float64                `json:"statistical_anomaly"`

	// Evidence
	SuspiciousEvents    []string               `json:"suspicious_events"`
	EvidenceDetails     map[string]interface{} `json:"evidence_details"`
	ConfidenceScore     float64                `json:"confidence_score"`
	RiskLevel           string                 `json:"risk_level"` // "low", "medium", "high", "critical"
	SuspiciousActivity  bool                   `json:"suspicious_activity"`
	CheatFlags          []string               `json:"cheat_flags"`

	// Actions taken
	FlaggedForReview    bool                   `json:"flagged_for_review"`
	ScoreInvalidated    bool                   `json:"score_invalidated"`
	LeaderboardExcluded bool                   `json:"leaderboard_excluded"`

	// Timestamps
	DetectedAt          time.Time              `json:"detected_at"`
	ReviewedAt          *time.Time             `json:"reviewed_at"`
	ReviewedBy          string                 `json:"reviewed_by"`
}

// AntiCheatConfig holds configuration for anti-cheat detection
type AntiCheatConfig struct {
	Enabled           bool               `json:"enabled"`
	StrictMode        bool               `json:"strict_mode"`
	VerificationRequired bool             `json:"verification_required"`
	VerificationMethods []string         `json:"verification_methods"`
	Thresholds        *DetectionThresholds `json:"thresholds"`
}

// DetectionThresholds holds threshold values for various cheat detection methods
type DetectionThresholds struct {
	MaxKPS                float64 `json:"max_kps"`
	MinBurstConsistency   float64 `json:"min_burst_consistency"`
	MaxErrorRate          float64 `json:"max_error_rate"`
	MinIdleTimePercent    float64 `json:"min_idle_time_percent"`
	MaxPasteEventRate     float64 `json:"max_paste_event_rate"`
	MinConfidenceScore    float64 `json:"min_confidence_score"`
	MaxAnomalousEvents    int     `json:"max_anomalous_events"`
	StatisticalThreshold  float64 `json:"statistical_threshold"`
}

// AntiCheatService handles anti-cheat detection and analysis
type AntiCheatService struct {
	dsClient    *datastore.Client
	config      *AntiCheatConfig
	thresholds  *DetectionThresholds
}

// NewAntiCheatService creates a new anti-cheat service
func NewAntiCheatService(dsClient *datastore.Client) *AntiCheatService {
	service := &AntiCheatService{
		dsClient: dsClient,
		config:   getDefaultAntiCheatConfig(),
		thresholds: &DetectionThresholds{
			MaxKPS:              12.0,
			MinBurstConsistency: 0.7,
			MaxErrorRate:        0.15,
			MinIdleTimePercent:  0.05,
			MaxPasteEventRate:   0.02,
			MinConfidenceScore:  0.6,
			MaxAnomalousEvents:  3,
			StatisticalThreshold: 2.5,
		},
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
			SessionID:      sessionID,
			ConfidenceScore: 1.0,
		}, nil
	}

	// Perform all detection methods
	report := &AntiCheatReport{
		SessionID: sessionID,
		UserID:    events[0].UserID,
	}

	// 1. Paste event detection
	report.PasteDetected = s.detectPasteEvents(events)
	if report.PasteDetected {
		report.SuspiciousEvents = append(report.SuspiciousEvents, "paste_detected")
		report.EvidenceDetails["paste_events"] = s.getPasteEventEvidence(events)
	}

	// 2. Unrealistic KPS detection
	kps := s.calculateKPS(events)
	report.UnrealisticKPS = kps > s.thresholds.MaxKPS
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

	// Store report if suspicious
	if report.SuspiciousActivity || len(report.SuspiciousEvents) > 0 {
		err = s.storeReport(ctx, report)
		if err != nil {
			log.Printf("Failed to store anti-cheat report: %v", err)
		}
	}

	return report, nil
}

// DetectPasteEvents identifies potential paste events in typing sessions
func (s *AntiCheatService) detectPasteEvents(events []*ScoringEvent) bool {
	if len(events) < 10 {
		return false
	}

	// Look for rapid consecutive keystrokes that are too fast for normal typing
	rapidBursts := 0
	windowSize := 5 // Check last 5 events

	for i := len(events) - windowSize; i < len(events); i++ {
		if i >= windowSize {
			// Check burst speed in this window
			burstKPS := s.calculateBurstKPS(events[i-windowSize : i+1])
			if burstKPS > s.thresholds.MaxKPS*1.5 { // 50% faster than max normal speed
				rapidBursts++
			}
		}
	}

	// If multiple rapid bursts detected, likely paste event
	return rapidBursts >= 2
}

// DetectAutoTypePattern identifies bot-like typing patterns
func (s *AntiCheatService) detectAutoTypePattern(events []*ScoringEvent) bool {
	if len(events) < 50 {
		return false
	}

	// Check for perfect rhythm consistency (bots often have exact timing)
	intervalVariance := s.calculateIntervalVariance(events)

	// Too perfect rhythm is suspicious
	if intervalVariance < 0.1 { // Less than 10% variance
		return true
	}

	// Check for lack of natural typing pauses
	pauseRatio := s.calculatePauseRatio(events)
	if pauseRatio < 0.05 { // Less than 5% pauses
		return true
	}

	return false
}

// DetectWindowFocusLoss detects if user switched away from typing window
func (s *AntiCheatService) detectWindowFocusLoss(events []*ScoringEvent) bool {
	// TODO: Implement window focus detection using metadata
	// This would require client-side focus/blur event tracking

	// For now, detect long gaps in typing that suggest window switching
	longGaps := 0
	for i := 1; i < len(events); i++ {
		gap := events[i].Timestamp - events[i-1].Timestamp
		if gap > 5000 { // Gap longer than 5 seconds
			longGaps++
		}
	}

	// Multiple long gaps suggest window focus loss
	return longGaps >= 3
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

	// Compare against historical baselines (would need user/session history)
	// For now, use hardcoded statistical thresholds
	deviations := 0

	if math.Abs(kps-6.0) > s.thresholds.StatisticalThreshold { // Average KPS around 6
		deviations++
	}

	if math.Abs(accuracy-0.95) > 0.1 { // Average accuracy around 95%
		deviations++
	}

	if math.Abs(errorRate-0.05) > 0.05 { // Average error rate around 5%
		deviations++
	}

	if burstConsistency < 0.8 { // Should have reasonable consistency
		deviations++
	}

	return deviations >= 2
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

func (s *AntiCheatService) calculateErrorRate(events []*ScoringEvent) float64 {
	errors := 0
	total := 0

	for _, event := range events {
		if event.EventType == "keystroke" && event.Action == "down" {
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

	return variance / (mean * mean) // Coefficient of variation
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
	deviations := 0.0

	kps := s.calculateKPS(events)
	accuracy := s.calculateAccuracy(events)

	// Z-score like calculation against expected values
	kpsDeviation := math.Abs(kps - 6.0) / 2.0  // Expected KPS around 6
	accuracyDeviation := math.Abs(accuracy - 0.95) / 0.1 // Expected accuracy around 95%

	deviations += kpsDeviation + accuracyDeviation

	return deviations
}

func (s *AntiCheatService) calculateOverallStatisticalAnomaly(events []*ScoringEvent) float64 {
	score := s.calculateStatisticalScore(events)

	// Check for multiple anomalies
	anomalies := 0
	if score > 2.0 {
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
	rapidBursts := 0
	for i := 0; i <= len(events)-5; i++ {
		burstKPS := s.calculateBurstKPS(events[i : i+5])
		if burstKPS > s.thresholds.MaxKPS*1.5 {
			rapidBursts++
		}
	}
	return map[string]interface{}{
		"rapid_bursts": rapidBursts,
		"threshold":    s.thresholds.MaxKPS * 1.5,
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
	// TODO: Query SessionEvent entities from Datastore
	// For now, return empty slice
	return []*ScoringEvent{}, nil
}

func (s *AntiCheatService) storeReport(ctx context.Context, report *AntiCheatReport) error {
	// Set timestamps
	report.DetectedAt = time.Now()

	// TODO: Store AntiCheatReport in Datastore
	key := datastore.NameKey("AntiCheatReport", fmt.Sprintf("%s_%d", report.SessionID, time.Now().Unix()), nil)

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
			MaxKPS:              12.0,
			MinBurstConsistency: 0.7,
			MaxErrorRate:        0.15,
			MinIdleTimePercent:  0.05,
			MaxPasteEventRate:   0.02,
			MinConfidenceScore:  0.6,
			MaxAnomalousEvents:  3,
			StatisticalThreshold: 2.5,
		},
	}
}
