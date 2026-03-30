package scoring

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
)

// MockDatastoreClient is a mock implementation of DatastoreClient for testing.
type MockDatastoreClient struct {
	mock.Mock
}

func (m *MockDatastoreClient) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	args := m.Called(ctx, key, src)
	return args.Get(0).(*datastore.Key), args.Error(1)
}

func (m *MockDatastoreClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	args := m.Called(ctx, key, dst)
	return args.Error(0)
}

func (m *MockDatastoreClient) Run(ctx context.Context, q *datastore.Query) Iterator {
	args := m.Called(ctx, q)
	return args.Get(0).(Iterator)
}

func (m *MockDatastoreClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	args := m.Called(ctx, q, dst)
	return args.Get(0).([]*datastore.Key), args.Error(1)
}

// MockIterator is a mock implementation of Iterator for testing.
type MockIterator struct {
	mock.Mock
}

func (m *MockIterator) Next(dst interface{}) (*datastore.Key, error) {
	args := m.Called(dst)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	return val.(*datastore.Key), args.Error(1)
}

func TestNewService(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewService(dsClient)

	assert.NotNil(t, service)
	assert.NotNil(t, service.dsClient)
	assert.NotNil(t, service.cache)
	assert.NotNil(t, service.eventQueue)
	assert.Equal(t, 4, service.workerCount)
	assert.NotNil(t, service.config)
}

func TestProcessEvent(t *testing.T) {
	dsClient := new(MockDatastoreClient)

	// Set up mock to return empty iterator for getSessionEvents calls
	mockIter := new(MockIterator)
	mockIter.On("Next", mock.Anything).Return(nil, iterator.Done)
	dsClient.On("Run", mock.Anything, mock.Anything).Return(mockIter)

	service := NewService(dsClient)
	ctx := context.Background()

	event := &ScoringEvent{
		SessionID:  "test-session-1",
		UserID:     "test-user-1",
		EventType:  "keystroke",
		Timestamp:  time.Now().UnixMilli(),
		KeyPressed: "a",
		Action:     "down",
		CursorPos:  0,
		ErrorFlag:  false,
	}

	// Test normal processing - should succeed
	err := service.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Test queue full by creating a service with a very small queue for testing
	smallQueueService := &Service{
		dsClient:    dsClient,
		cache:       &sync.Map{},
		eventQueue:  make(chan *ScoringEvent, 1), // Only 1 item buffer
		workerCount: 0,                           // No workers to avoid processing
		config:      service.getDefaultConfig(),
	}

	// Fill the small queue
	err = smallQueueService.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Try to add another - should fail
	err = smallQueueService.ProcessEvent(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event queue full")
}

func TestAnalyzeSession(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)
	ctx := context.Background()

	// Mock getSessionEvents to return some events
	events := []*ScoringEvent{
		{UserID: "user1", Timestamp: 1000, EventType: "keystroke", KeyPressed: "a", Action: "down", ErrorFlag: false},
		{UserID: "user1", Timestamp: 1100, EventType: "keystroke", KeyPressed: "s", Action: "down", ErrorFlag: false},
		{UserID: "user1", Timestamp: 1200, EventType: "keystroke", KeyPressed: "d", Action: "down", ErrorFlag: true},
		{UserID: "user1", Timestamp: 1300, EventType: "keystroke", KeyPressed: "f", Action: "down", ErrorFlag: false},
	}

	dsClient.On("Run", mock.Anything, mock.Anything).Return(func(ctx context.Context, q *datastore.Query) Iterator {
		iter := new(MockIterator)
		callCount := 0
		iter.On("Next", mock.Anything).Return(func(dst interface{}) (*datastore.Key, error) {
			if callCount < len(events) {
				*dst.(*ScoringEvent) = *events[callCount]
				callCount++
				return datastore.IncompleteKey("ScoringEvent", nil), nil
			}
			return nil, errors.New("done")
		}).Times(len(events) + 1)
		return iter
	}).Once()

	dsClient.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(datastore.IncompleteKey("AntiCheatReport", nil), nil).Once()

	report, err := service.AnalyzeSession(ctx, "session123")
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.False(t, report.PasteDetected)
	assert.False(t, report.UnrealisticKPS)
	assert.False(t, report.AutoTypePattern)
	assert.False(t, report.WindowFocusLost)
	assert.False(t, report.AnomalousTiming)
	assert.Equal(t, "low", report.RiskLevel)
	assert.False(t, report.SuspiciousActivity)

	dsClient.AssertExpectations(t)
}

func TestDetectPasteEvents(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	// Test case: no paste events
	events := []*ScoringEvent{
		{Timestamp: 1000, EventType: "keystroke"},
		{Timestamp: 1200, EventType: "keystroke"},
		{Timestamp: 1400, EventType: "keystroke"},
		{Timestamp: 1600, EventType: "keystroke"},
		{Timestamp: 1800, EventType: "keystroke"},
		{Timestamp: 2000, EventType: "keystroke"},
		{Timestamp: 2200, EventType: "keystroke"},
		{Timestamp: 2400, EventType: "keystroke"},
		{Timestamp: 2600, EventType: "keystroke"},
		{Timestamp: 2800, EventType: "keystroke"},
		{Timestamp: 3000, EventType: "keystroke"},
	}
	assert.False(t, service.detectPasteEvents(events))

	// Test case: with paste events (rapid bursts)
	eventsWithPaste := []*ScoringEvent{
		{Timestamp: 1000, EventType: "keystroke"},
		{Timestamp: 1010, EventType: "keystroke"},
		{Timestamp: 1020, EventType: "keystroke"},
		{Timestamp: 1030, EventType: "keystroke"},
		{Timestamp: 1040, EventType: "keystroke"},
		{Timestamp: 1050, EventType: "keystroke"},
		{Timestamp: 1060, EventType: "keystroke"},
		{Timestamp: 1070, EventType: "keystroke"},
		{Timestamp: 1080, EventType: "keystroke"},
		{Timestamp: 1090, EventType: "keystroke"},
		{Timestamp: 1100, EventType: "keystroke"},
		{Timestamp: 1110, EventType: "keystroke"},
		{Timestamp: 1120, EventType: "keystroke"},
		{Timestamp: 1130, EventType: "keystroke"},
		{Timestamp: 1140, EventType: "keystroke"},
		{Timestamp: 1150, EventType: "keystroke"},
		{Timestamp: 1160, EventType: "keystroke"},
		{Timestamp: 1170, EventType: "keystroke"},
		{Timestamp: 1180, EventType: "keystroke"},
		{Timestamp: 1190, EventType: "keystroke"},
		{Timestamp: 1200, EventType: "keystroke"},
	}
	assert.True(t, service.detectPasteEvents(eventsWithPaste))
}

func TestDetectAutoTypePattern(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	// Test case: no auto-type pattern
	events := make([]*ScoringEvent, 60)
	for i := 0; i < 60; i++ {
		events[i] = &ScoringEvent{Timestamp: int64(1000 + i*150), EventType: "keystroke"}
	}
	assert.False(t, service.detectAutoTypePattern(events))

	// Test case: auto-type pattern (very consistent intervals)
	eventsAutoType := make([]*ScoringEvent, 60)
	for i := 0; i < 60; i++ {
		eventsAutoType[i] = &ScoringEvent{Timestamp: int64(1000 + i*100), EventType: "keystroke"}
	}
	assert.True(t, service.detectAutoTypePattern(eventsAutoType))

	// Test case: auto-type pattern (very low pause ratio)
	eventsLowPause := make([]*ScoringEvent, 60)
	for i := 0; i < 60; i++ {
		eventsLowPause[i] = &ScoringEvent{Timestamp: int64(1000 + i*50), EventType: "keystroke"}
	}
	assert.True(t, service.detectAutoTypePattern(eventsLowPause))
}

func TestDetectWindowFocusLoss(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	// Test case: no window focus loss
	events := []*ScoringEvent{
		{Timestamp: 1000},
		{Timestamp: 2000},
		{Timestamp: 3000},
		{Timestamp: 4000},
		{Timestamp: 5000},
	}
	assert.False(t, service.detectWindowFocusLoss(events))

	// Test case: with window focus loss (long gaps)
	eventsWithLoss := []*ScoringEvent{
		{Timestamp: 1000},
		{Timestamp: 7000}, // 6s gap
		{Timestamp: 8000},
		{Timestamp: 14000}, // 6s gap
		{Timestamp: 15000},
		{Timestamp: 21000}, // 6s gap
		{Timestamp: 22000},
	}
	assert.True(t, service.detectWindowFocusLoss(eventsWithLoss))
}

func TestDetectStatisticalAnomalies(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	// Test case: no anomalies
	events := make([]*ScoringEvent, 30)
	for i := 0; i < 30; i++ {
		events[i] = &ScoringEvent{Timestamp: int64(1000 + i*100), EventType: "keystroke", ErrorFlag: false}
	}
	assert.False(t, service.detectStatisticalAnomalies(events))

	// Test case: with anomalies (very high KPS)
	eventsHighKPS := make([]*ScoringEvent, 30)
	for i := 0; i < 30; i++ {
		eventsHighKPS[i] = &ScoringEvent{Timestamp: int64(1000 + i*10), EventType: "keystroke", ErrorFlag: false}
	}
	assert.True(t, service.detectStatisticalAnomalies(eventsHighKPS))

	// Test case: with anomalies (very low accuracy)
	eventsLowAccuracy := make([]*ScoringEvent, 30)
	for i := 0; i < 30; i++ {
		eventsLowAccuracy[i] = &ScoringEvent{Timestamp: int64(1000 + i*100), EventType: "keystroke", ErrorFlag: true}
	}
	assert.True(t, service.detectStatisticalAnomalies(eventsLowAccuracy))
}

func TestCalculateKPS(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	events := []*ScoringEvent{
		{Timestamp: 1000, EventType: "keystroke"},
		{Timestamp: 2000, EventType: "keystroke"},
		{Timestamp: 3000, EventType: "keystroke"},
		{Timestamp: 4000, EventType: "keystroke"},
		{Timestamp: 5000, EventType: "keystroke"},
	}
	kps := service.calculateKPS(events)
	assert.InDelta(t, 1.25, kps, 0.01)

	// Edge case: single event
	kps = service.calculateKPS([]*ScoringEvent{{Timestamp: 1000, EventType: "keystroke"}})
	assert.Equal(t, 0.0, kps)

	// Edge case: no events
	kps = service.calculateKPS([]*ScoringEvent{})
	assert.Equal(t, 0.0, kps)
}

func TestCalculateAccuracy(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	events := []*ScoringEvent{
		{EventType: "keystroke", Action: "down", ErrorFlag: false},
		{EventType: "keystroke", Action: "down", ErrorFlag: false},
		{EventType: "keystroke", Action: "down", ErrorFlag: true},
		{EventType: "keystroke", Action: "down", ErrorFlag: false},
	}
	accuracy := service.calculateAccuracy(events)
	assert.InDelta(t, 0.75, accuracy, 0.01)

	// Edge case: no events
	accuracy = service.calculateAccuracy([]*ScoringEvent{})
	assert.Equal(t, 0.0, accuracy)
}

func TestCalculateErrorRate(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	events := []*ScoringEvent{
		{EventType: "keystroke", Action: "down", ErrorFlag: false},
		{EventType: "keystroke", Action: "down", ErrorFlag: false},
		{EventType: "keystroke", Action: "down", ErrorFlag: true},
		{EventType: "keystroke", Action: "down", ErrorFlag: false},
	}
	errorRate := service.calculateErrorRate(events)
	assert.InDelta(t, 0.25, errorRate, 0.01)

	// Edge case: no events
	errorRate = service.calculateErrorRate([]*ScoringEvent{})
	assert.Equal(t, 0.0, errorRate)
}

func TestCalculateBurstKPS(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	events := []*ScoringEvent{
		{Timestamp: 1000, EventType: "keystroke"},
		{Timestamp: 1100, EventType: "keystroke"},
		{Timestamp: 1200, EventType: "keystroke"},
		{Timestamp: 1300, EventType: "keystroke"},
		{Timestamp: 1400, EventType: "keystroke"},
	}
	burstKPS := service.calculateBurstKPS(events)
	assert.InDelta(t, 12.5, burstKPS, 0.01)

	// Edge case: less than 2 events
	burstKPS = service.calculateBurstKPS([]*ScoringEvent{{Timestamp: 1000, EventType: "keystroke"}})
	assert.Equal(t, 0.0, burstKPS)
}

func TestCalculateIntervalVariance(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	events := []*ScoringEvent{
		{Timestamp: 1000, EventType: "keystroke"},
		{Timestamp: 1100, EventType: "keystroke"},
		{Timestamp: 1200, EventType: "keystroke"},
		{Timestamp: 1300, EventType: "keystroke"},
		{Timestamp: 1400, EventType: "keystroke"},
	}
	variance := service.calculateIntervalVariance(events)
	assert.InDelta(t, 0.0, variance, 0.01)

	// Test with some variance
	eventsWithVariance := []*ScoringEvent{
		{Timestamp: 1000, EventType: "keystroke"},
		{Timestamp: 1100, EventType: "keystroke"},
		{Timestamp: 1300, EventType: "keystroke"},
		{Timestamp: 1400, EventType: "keystroke"},
		{Timestamp: 1700, EventType: "keystroke"},
	}
	variance = service.calculateIntervalVariance(eventsWithVariance)
	assert.InDelta(t, 0.75, variance, 0.01)
}

func TestCalculatePauseRatio(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	// No pauses
	events := []*ScoringEvent{
		{Timestamp: 1000},
		{Timestamp: 1100},
		{Timestamp: 1200},
		{Timestamp: 1300},
		{Timestamp: 1400},
	}
	assert.Equal(t, 0.0, service.calculatePauseRatio(events))

	// With pauses
	eventsWithPauses := []*ScoringEvent{
		{Timestamp: 1000},
		{Timestamp: 1500}, // 500ms pause
		{Timestamp: 1600},
		{Timestamp: 2200}, // 600ms pause
		{Timestamp: 2300},
	}
	// Total duration: 1300ms, total pause time: 1100ms
	assert.InDelta(t, 1100.0/1300.0, service.calculatePauseRatio(eventsWithPauses), 0.01)
}

func TestCalculateRiskLevel(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	assert.Equal(t, "low", service.calculateRiskLevel(0))
	assert.Equal(t, "medium", service.calculateRiskLevel(1))
	assert.Equal(t, "high", service.calculateRiskLevel(2))
	assert.Equal(t, "critical", service.calculateRiskLevel(3))
	assert.Equal(t, "critical", service.calculateRiskLevel(5))
}

func TestCalculateConfidenceScore(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	// No suspicious events, short session
	events := make([]*ScoringEvent, 10)
	score := service.calculateConfidenceScore(events, []string{})
	assert.InDelta(t, 0.5, score, 0.01)

	// One suspicious event, short session
	score = service.calculateConfidenceScore(events, []string{"paste_detected"})
	assert.InDelta(t, 0.7, score, 0.01)

	// Two suspicious events, long session
	eventsLong := make([]*ScoringEvent, 150)
	score = service.calculateConfidenceScore(eventsLong, []string{"paste_detected", "unrealistic_kps"})
	assert.InDelta(t, 1.0, score, 0.01)
}

func TestDetermineActions(t *testing.T) {
	dsClient := new(MockDatastoreClient)
	service := NewAntiCheatService(dsClient)

	// Critical risk
	report := &AntiCheatReport{RiskLevel: "critical"}
	service.determineActions(report)
	assert.True(t, report.FlaggedForReview)
	assert.True(t, report.ScoreInvalidated)
	assert.True(t, report.LeaderboardExcluded)

	// High risk, high confidence
	report = &AntiCheatReport{RiskLevel: "high", ConfidenceScore: 0.9}
	service.determineActions(report)
	assert.True(t, report.FlaggedForReview)
	assert.True(t, report.ScoreInvalidated)
	assert.True(t, report.LeaderboardExcluded)

	// High risk, low confidence
	report = &AntiCheatReport{RiskLevel: "high", ConfidenceScore: 0.5}
	service.determineActions(report)
	assert.True(t, report.FlaggedForReview)
	assert.False(t, report.ScoreInvalidated)
	assert.False(t, report.LeaderboardExcluded)

	// Medium risk
	report = &AntiCheatReport{RiskLevel: "medium", ConfidenceScore: 0.7}
	service.determineActions(report)
	assert.True(t, report.FlaggedForReview)
	assert.False(t, report.ScoreInvalidated)
	assert.False(t, report.LeaderboardExcluded)

	// Low risk
	report = &AntiCheatReport{RiskLevel: "low"}
	service.determineActions(report)
	assert.False(t, report.FlaggedForReview)
	assert.False(t, report.ScoreInvalidated)
	assert.False(t, report.LeaderboardExcluded)
}

// MockLeaderboardService is a mock implementation of the LeaderboardService interface.
type MockLeaderboardService struct {
	GetLeaderboardFunc       func(ctx context.Context, req *LeaderboardRequest) (*LeaderboardResponse, error)
	UpdateLeaderboardFunc    func(ctx context.Context, sessionID string) error
	GetUserRankFunc          func(ctx context.Context, userID, languageID, mode, scope, timeWindow string) (*UserRank, error)
	GetLeaderboardTrendsFunc func(ctx context.Context, req *LeaderboardTrendsRequest) (*LeaderboardTrendsResponse, error)
}

// GetLeaderboard implements LeaderboardService.
func (m *MockLeaderboardService) GetLeaderboard(ctx context.Context, req *LeaderboardRequest) (*LeaderboardResponse, error) {
	if m.GetLeaderboardFunc != nil {
		return m.GetLeaderboardFunc(ctx, req)
	}
	return nil, nil
}

// UpdateLeaderboard implements LeaderboardService.
func (m *MockLeaderboardService) UpdateLeaderboard(ctx context.Context, sessionID string) error {
	if m.UpdateLeaderboardFunc != nil {
		return m.UpdateLeaderboardFunc(ctx, sessionID)
	}
	return nil
}

// GetUserRank implements LeaderboardService.
func (m *MockLeaderboardService) GetUserRank(ctx context.Context, userID, languageID, mode, scope, timeWindow string) (*UserRank, error) {
	if m.GetUserRankFunc != nil {
		return m.GetUserRankFunc(ctx, userID, languageID, mode, scope, timeWindow)
	}
	return nil, nil
}

// GetLeaderboardTrends implements LeaderboardService.
func (m *MockLeaderboardService) GetLeaderboardTrends(ctx context.Context, req *LeaderboardTrendsRequest) (*LeaderboardTrendsResponse, error) {
	if m.GetLeaderboardTrendsFunc != nil {
		return m.GetLeaderboardTrendsFunc(ctx, req)
	}
	return nil, nil
}
