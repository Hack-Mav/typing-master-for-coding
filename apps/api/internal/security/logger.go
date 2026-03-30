package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
)

// SecurityLogger handles security event logging and alerting
type SecurityLogger struct {
	db     *database.DatastoreClient
	config *LoggerConfig
}

// LoggerConfig holds configuration for security logging
type LoggerConfig struct {
	EnableFileLogging     bool
	EnableDatabaseLogging bool
	EnableAlerting        bool
	AlertThreshold        int
	AlertWindow           time.Duration
	LogLevel              string
}

// SecurityEvent represents a security event for logging
type SecurityEvent struct {
	ID          string                 `datastore:"-" json:"id"`
	Type        string                 `datastore:"type" json:"type"`
	Severity    string                 `datastore:"severity" json:"severity"`
	Description string                 `datastore:"description" json:"description"`
	Source      string                 `datastore:"source" json:"source"`
	UserAgent   string                 `datastore:"user_agent" json:"user_agent"`
	IP          string                 `datastore:"ip" json:"ip"`
	UserID      string                 `datastore:"user_id,omitempty" json:"user_id,omitempty"`
	SessionID   string                 `datastore:"session_id,omitempty" json:"session_id,omitempty"`
	Metadata    map[string]interface{} `datastore:"metadata,omitempty" json:"metadata,omitempty"`
	Blocked     bool                   `datastore:"blocked" json:"blocked"`
	Timestamp   time.Time              `datastore:"timestamp" json:"timestamp"`
	CreatedAt   time.Time              `datastore:"created_at" json:"created_at"`
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string    `datastore:"-" json:"id"`
	Type        string    `datastore:"type" json:"type"`
	Severity    string    `datastore:"severity" json:"severity"`
	Title       string    `datastore:"title" json:"title"`
	Description string    `datastore:"description" json:"description"`
	EventCount  int       `datastore:"event_count" json:"event_count"`
	FirstSeen   time.Time `datastore:"first_seen" json:"first_seen"`
	LastSeen    time.Time `datastore:"last_seen" json:"last_seen"`
	Resolved    bool      `datastore:"resolved" json:"resolved"`
	ResolvedBy  string    `datastore:"resolved_by,omitempty" json:"resolved_by,omitempty"`
	ResolvedAt  time.Time `datastore:"resolved_at,omitempty" json:"resolved_at,omitempty"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

// NewSecurityLogger creates a new security logger
func NewSecurityLogger(db *database.DatastoreClient, config *LoggerConfig) *SecurityLogger {
	if config == nil {
		config = DefaultLoggerConfig()
	}
	return &SecurityLogger{
		db:     db,
		config: config,
	}
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		EnableFileLogging:     true,
		EnableDatabaseLogging: true,
		EnableAlerting:        true,
		AlertThreshold:        10,
		AlertWindow:           time.Hour,
		LogLevel:              "INFO",
	}
}

// LogSecurityEvent logs a security event
func (sl *SecurityLogger) LogSecurityEvent(ctx context.Context, threat SecurityThreat) error {
	event := SecurityEvent{
		Type:        threat.Type,
		Severity:    threat.Severity,
		Description: threat.Description,
		Source:      threat.Source,
		UserAgent:   threat.UserAgent,
		IP:          threat.IP,
		Blocked:     threat.Blocked,
		Timestamp:   threat.Timestamp,
		CreatedAt:   time.Now().UTC(),
	}

	// Log to file/console
	if sl.config.EnableFileLogging {
		sl.logToFile(event)
	}

	// Log to database
	if sl.config.EnableDatabaseLogging && sl.db != nil {
		if err := sl.logToDatabase(ctx, event); err != nil {
			log.Printf("Failed to log security event to database: %v", err)
		}
	}

	// Check for alerting
	if sl.config.EnableAlerting {
		if err := sl.checkAndCreateAlert(ctx, event); err != nil {
			log.Printf("Failed to check security alerts: %v", err)
		}
	}

	return nil
}

// logToFile logs security event to file/console
func (sl *SecurityLogger) logToFile(event SecurityEvent) {
	eventJSON, _ := json.Marshal(event)
	log.Printf("SECURITY_EVENT: %s", string(eventJSON))
}

// logToDatabase logs security event to database
func (sl *SecurityLogger) logToDatabase(ctx context.Context, event SecurityEvent) error {
	if sl.db == nil {
		return fmt.Errorf("database client not initialized")
	}

	eventID := uuid.New().String()
	key := datastore.NameKey("SecurityEvent", eventID, nil)

	_, err := sl.db.Put(ctx, key, &event)
	if err != nil {
		return fmt.Errorf("failed to save security event: %w", err)
	}

	return nil
}

// checkAndCreateAlert checks if an alert should be created based on event patterns
func (sl *SecurityLogger) checkAndCreateAlert(ctx context.Context, event SecurityEvent) error {
	if sl.db == nil {
		return nil // Skip alerting if no database
	}

	// Count recent events of the same type from the same IP
	now := time.Now().UTC()
	windowStart := now.Add(-sl.config.AlertWindow)

	query := datastore.NewQuery("SecurityEvent").
		Filter("type =", event.Type).
		Filter("ip =", event.IP).
		Filter("timestamp >=", windowStart)

	var events []SecurityEvent
	_, err := sl.db.GetAll(ctx, query, &events)
	if err != nil {
		return fmt.Errorf("failed to query recent events: %w", err)
	}

	// If threshold exceeded, create or update alert
	if len(events) >= sl.config.AlertThreshold {
		return sl.createOrUpdateAlert(ctx, event, len(events))
	}

	return nil
}

// createOrUpdateAlert creates a new alert or updates an existing one
func (sl *SecurityLogger) createOrUpdateAlert(ctx context.Context, event SecurityEvent, eventCount int) error {
	// Check if there's an existing unresolved alert for this type and IP
	query := datastore.NewQuery("SecurityAlert").
		Filter("type =", event.Type).
		Filter("resolved =", false).
		Limit(1)

	var alerts []SecurityAlert
	keys, err := sl.db.GetAll(ctx, query, &alerts)
	if err != nil {
		return fmt.Errorf("failed to query existing alerts: %w", err)
	}

	now := time.Now().UTC()

	if len(alerts) > 0 {
		// Update existing alert
		alert := alerts[0]
		alert.EventCount = eventCount
		alert.LastSeen = now
		alert.UpdatedAt = now

		key := keys[0]
		_, err = sl.db.Put(ctx, key, &alert)
		if err != nil {
			return fmt.Errorf("failed to update security alert: %w", err)
		}
	} else {
		// Create new alert
		alert := SecurityAlert{
			Type:        event.Type,
			Severity:    event.Severity,
			Title:       fmt.Sprintf("Multiple %s attempts detected", event.Type),
			Description: fmt.Sprintf("Detected %d %s attempts from IP %s in the last %v", eventCount, event.Type, event.IP, sl.config.AlertWindow),
			EventCount:  eventCount,
			FirstSeen:   now,
			LastSeen:    now,
			Resolved:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		alertID := uuid.New().String()
		key := datastore.NameKey("SecurityAlert", alertID, nil)
		_, err = sl.db.Put(ctx, key, &alert)
		if err != nil {
			return fmt.Errorf("failed to create security alert: %w", err)
		}

		// Log the alert creation
		log.Printf("SECURITY_ALERT_CREATED: %s - %s", alert.Type, alert.Description)
	}

	return nil
}

// GetSecurityEvents retrieves security events with filtering
func (sl *SecurityLogger) GetSecurityEvents(ctx context.Context, filters SecurityEventFilters) ([]SecurityEvent, error) {
	query := datastore.NewQuery("SecurityEvent")

	// Apply filters
	if filters.Type != "" {
		query = query.Filter("type =", filters.Type)
	}
	if filters.Severity != "" {
		query = query.Filter("severity =", filters.Severity)
	}
	if filters.IP != "" {
		query = query.Filter("ip =", filters.IP)
	}
	if !filters.StartTime.IsZero() {
		query = query.Filter("timestamp >=", filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		query = query.Filter("timestamp <=", filters.EndTime)
	}

	// Apply ordering and limit
	query = query.Order("-timestamp")
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	var events []SecurityEvent
	keys, err := sl.db.GetAll(ctx, query, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	for i, key := range keys {
		events[i].ID = key.Name
	}

	return events, nil
}

// GetSecurityAlerts retrieves security alerts
func (sl *SecurityLogger) GetSecurityAlerts(ctx context.Context, includeResolved bool) ([]SecurityAlert, error) {
	query := datastore.NewQuery("SecurityAlert")

	if !includeResolved {
		query = query.Filter("resolved =", false)
	}

	query = query.Order("-created_at")

	var alerts []SecurityAlert
	keys, err := sl.db.GetAll(ctx, query, &alerts)
	if err != nil {
		return nil, fmt.Errorf("failed to get security alerts: %w", err)
	}

	for i, key := range keys {
		alerts[i].ID = key.Name
	}

	return alerts, nil
}

// ResolveAlert marks a security alert as resolved
func (sl *SecurityLogger) ResolveAlert(ctx context.Context, alertID, resolvedBy string) error {
	key := datastore.NameKey("SecurityAlert", alertID, nil)
	var alert SecurityAlert
	err := sl.db.Get(ctx, key, &alert)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alert.Resolved = true
	alert.ResolvedBy = resolvedBy
	alert.ResolvedAt = time.Now().UTC()
	alert.UpdatedAt = time.Now().UTC()

	_, err = sl.db.Put(ctx, key, &alert)
	if err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	return nil
}

// SecurityEventFilters holds filters for querying security events
type SecurityEventFilters struct {
	Type      string
	Severity  string
	IP        string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// GetSecurityMetrics returns security metrics for monitoring
func (sl *SecurityLogger) GetSecurityMetrics(ctx context.Context, window time.Duration) (*SecurityMetrics, error) {
	now := time.Now().UTC()
	windowStart := now.Add(-window)

	// Query events in the time window
	query := datastore.NewQuery("SecurityEvent").
		Filter("timestamp >=", windowStart)

	var events []SecurityEvent
	_, err := sl.db.GetAll(ctx, query, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	// Calculate metrics
	metrics := &SecurityMetrics{
		TotalEvents:      len(events),
		BlockedEvents:    0,
		EventsByType:     make(map[string]int),
		EventsBySeverity: make(map[string]int),
		TopIPs:           make(map[string]int),
		WindowStart:      windowStart,
		WindowEnd:        now,
	}

	for _, event := range events {
		if event.Blocked {
			metrics.BlockedEvents++
		}
		metrics.EventsByType[event.Type]++
		metrics.EventsBySeverity[event.Severity]++
		metrics.TopIPs[event.IP]++
	}

	return metrics, nil
}

// SecurityMetrics holds security metrics
type SecurityMetrics struct {
	TotalEvents      int            `json:"total_events"`
	BlockedEvents    int            `json:"blocked_events"`
	EventsByType     map[string]int `json:"events_by_type"`
	EventsBySeverity map[string]int `json:"events_by_severity"`
	TopIPs           map[string]int `json:"top_ips"`
	WindowStart      time.Time      `json:"window_start"`
	WindowEnd        time.Time      `json:"window_end"`
}
