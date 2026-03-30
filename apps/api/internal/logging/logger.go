package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	ID         string                 `datastore:"-" json:"id"`
	Level      string                 `datastore:"level" json:"level"`
	Message    string                 `datastore:"message" json:"message"`
	Service    string                 `datastore:"service" json:"service"`
	Component  string                 `datastore:"component" json:"component"`
	UserID     string                 `datastore:"user_id,omitempty" json:"user_id,omitempty"`
	SessionID  string                 `datastore:"session_id,omitempty" json:"session_id,omitempty"`
	RequestID  string                 `datastore:"request_id,omitempty" json:"request_id,omitempty"`
	TraceID    string                 `datastore:"trace_id,omitempty" json:"trace_id,omitempty"`
	Error      string                 `datastore:"error,omitempty" json:"error,omitempty"`
	StackTrace string                 `datastore:"stack_trace,omitempty" json:"stack_trace,omitempty"`
	Metadata   map[string]interface{} `datastore:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp  time.Time              `datastore:"timestamp" json:"timestamp"`
	CreatedAt  time.Time              `datastore:"created_at" json:"created_at"`
}

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	ID        string                 `datastore:"-" json:"id"`
	Name      string                 `datastore:"name" json:"name"`
	Value     float64                `datastore:"value" json:"value"`
	Unit      string                 `datastore:"unit" json:"unit"`
	Service   string                 `datastore:"service" json:"service"`
	Component string                 `datastore:"component" json:"component"`
	Tags      map[string]interface{} `datastore:"tags,omitempty" json:"tags,omitempty"`
	Timestamp time.Time              `datastore:"timestamp" json:"timestamp"`
	CreatedAt time.Time              `datastore:"created_at" json:"created_at"`
}

// ErrorAlert represents an error alert for monitoring
type ErrorAlert struct {
	ID          string    `datastore:"-" json:"id"`
	Type        string    `datastore:"type" json:"type"`
	Severity    string    `datastore:"severity" json:"severity"`
	Title       string    `datastore:"title" json:"title"`
	Description string    `datastore:"description" json:"description"`
	Service     string    `datastore:"service" json:"service"`
	Component   string    `datastore:"component" json:"component"`
	ErrorCount  int       `datastore:"error_count" json:"error_count"`
	FirstSeen   time.Time `datastore:"first_seen" json:"first_seen"`
	LastSeen    time.Time `datastore:"last_seen" json:"last_seen"`
	Resolved    bool      `datastore:"resolved" json:"resolved"`
	ResolvedBy  string    `datastore:"resolved_by,omitempty" json:"resolved_by,omitempty"`
	ResolvedAt  time.Time `datastore:"resolved_at,omitempty" json:"resolved_at,omitempty"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

// Logger provides structured logging capabilities
type Logger struct {
	db             *database.DatastoreClient
	config         *LoggerConfig
	service        string
	component      string
	enableConsole  bool
	enableDatabase bool
	enableAlerting bool
	alertThreshold int
	alertWindow    time.Duration
	errorCounts    map[string]int
	lastErrorCheck time.Time
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Service          string        `json:"service"`
	Component        string        `json:"component"`
	MinLevel         LogLevel      `json:"min_level"`
	EnableConsole    bool          `json:"enable_console"`
	EnableDatabase   bool          `json:"enable_database"`
	EnableAlerting   bool          `json:"enable_alerting"`
	AlertThreshold   int           `json:"alert_threshold"`
	AlertWindow      time.Duration `json:"alert_window"`
	MaxRetentionDays int           `json:"max_retention_days"`
	BatchSize        int           `json:"batch_size"`
	FlushInterval    time.Duration `json:"flush_interval"`
}

// NewLogger creates a new structured logger
func NewLogger(db *database.DatastoreClient, config *LoggerConfig) *Logger {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	return &Logger{
		db:             db,
		config:         config,
		service:        config.Service,
		component:      config.Component,
		enableConsole:  config.EnableConsole,
		enableDatabase: config.EnableDatabase,
		enableAlerting: config.EnableAlerting,
		alertThreshold: config.AlertThreshold,
		alertWindow:    config.AlertWindow,
		errorCounts:    make(map[string]int),
		lastErrorCheck: time.Now(),
	}
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Service:          "backend/backend",
		Component:        "main",
		MinLevel:         INFO,
		EnableConsole:    true,
		EnableDatabase:   true,
		EnableAlerting:   true,
		AlertThreshold:   10,
		AlertWindow:      time.Hour,
		MaxRetentionDays: 30,
		BatchSize:        100,
		FlushInterval:    time.Minute * 5,
	}
}

// WithContext creates a new logger with context information
func (l *Logger) WithContext(ctx context.Context) *Logger {
	newLogger := *l

	// Extract context values if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			newLogger.component = fmt.Sprintf("%s[%s]", l.component, id)
		}
	}

	return &newLogger
}

// WithComponent creates a new logger with a specific component
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := *l
	newLogger.component = component
	return &newLogger
}

// Debug logs a debug message
func (l *Logger) Debug(message string, metadata ...map[string]interface{}) {
	l.log(DEBUG, message, nil, metadata...)
}

// Info logs an info message
func (l *Logger) Info(message string, metadata ...map[string]interface{}) {
	l.log(INFO, message, nil, metadata...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, metadata ...map[string]interface{}) {
	l.log(WARN, message, nil, metadata...)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, metadata ...map[string]interface{}) {
	l.log(ERROR, message, err, metadata...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, metadata ...map[string]interface{}) {
	l.log(FATAL, message, err, metadata...)
	os.Exit(1)
}

// LogSessionEvent logs a session-related event
func (l *Logger) LogSessionEvent(ctx context.Context, sessionID, userID, event string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["session_id"] = sessionID
	metadata["user_id"] = userID
	metadata["event_type"] = "session"

	l.log(INFO, fmt.Sprintf("Session event: %s", event), nil, metadata)
}

// LogPerformanceMetric logs a performance metric
func (l *Logger) LogPerformanceMetric(ctx context.Context, name string, value float64, unit string, tags map[string]interface{}) {
	metric := PerformanceMetric{
		Name:      name,
		Value:     value,
		Unit:      unit,
		Service:   l.service,
		Component: l.component,
		Tags:      tags,
		Timestamp: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	// Log to console if enabled
	if l.enableConsole {
		metricJSON, _ := json.Marshal(metric)
		log.Printf("PERFORMANCE_METRIC: %s", string(metricJSON))
	}

	// Store in database if enabled
	if l.enableDatabase && l.db != nil {
		go func() {
			if err := l.storePerformanceMetric(ctx, metric); err != nil {
				log.Printf("Failed to store performance metric: %v", err)
			}
		}()
	}
}

// LogUserFriendlyError logs an error with a user-friendly message
func (l *Logger) LogUserFriendlyError(ctx context.Context, userMessage, technicalMessage string, err error, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["user_message"] = userMessage
	metadata["technical_message"] = technicalMessage
	metadata["error_type"] = "user_facing"

	l.log(ERROR, userMessage, err, metadata)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string, err error, metadata ...map[string]interface{}) {
	// Check if we should log this level
	if !l.shouldLog(level) {
		return
	}

	now := time.Now().UTC()

	// Merge metadata
	mergedMetadata := make(map[string]interface{})
	for _, m := range metadata {
		for k, v := range m {
			mergedMetadata[k] = v
		}
	}

	entry := LogEntry{
		Level:     string(level),
		Message:   message,
		Service:   l.service,
		Component: l.component,
		Metadata:  mergedMetadata,
		Timestamp: now,
		CreatedAt: now,
	}

	// Add error information if provided
	if err != nil {
		entry.Error = err.Error()
		if level == ERROR || level == FATAL {
			entry.StackTrace = l.getStackTrace()
		}
	}

	// Extract context information from metadata
	if userID, ok := mergedMetadata["user_id"].(string); ok {
		entry.UserID = userID
	}
	if sessionID, ok := mergedMetadata["session_id"].(string); ok {
		entry.SessionID = sessionID
	}
	if requestID, ok := mergedMetadata["request_id"].(string); ok {
		entry.RequestID = requestID
	}
	if traceID, ok := mergedMetadata["trace_id"].(string); ok {
		entry.TraceID = traceID
	}

	// Log to console if enabled
	if l.enableConsole {
		l.logToConsole(entry)
	}

	// Store in database if enabled
	if l.enableDatabase && l.db != nil {
		go func() {
			if err := l.storeLogEntry(context.Background(), entry); err != nil {
				log.Printf("Failed to store log entry: %v", err)
			}
		}()
	}

	// Check for alerting if this is an error
	if l.enableAlerting && (level == ERROR || level == FATAL) {
		go func() {
			if err := l.checkAndCreateAlert(context.Background(), entry); err != nil {
				log.Printf("Failed to check alerts: %v", err)
			}
		}()
	}
}

// shouldLog checks if the log level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
		FATAL: 4,
	}

	return levels[level] >= levels[l.config.MinLevel]
}

// logToConsole outputs the log entry to console
func (l *Logger) logToConsole(entry LogEntry) {
	// Format for console output
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")

	var output string
	if entry.Error != "" {
		output = fmt.Sprintf("[%s] %s [%s/%s] %s - Error: %s",
			timestamp, entry.Level, entry.Service, entry.Component, entry.Message, entry.Error)
	} else {
		output = fmt.Sprintf("[%s] %s [%s/%s] %s",
			timestamp, entry.Level, entry.Service, entry.Component, entry.Message)
	}

	// Add metadata if present
	if len(entry.Metadata) > 0 {
		metadataJSON, _ := json.Marshal(entry.Metadata)
		output += fmt.Sprintf(" | Metadata: %s", string(metadataJSON))
	}

	log.Println(output)
}

// storeLogEntry stores the log entry in the database
func (l *Logger) storeLogEntry(ctx context.Context, entry LogEntry) error {
	entryID := uuid.New().String()
	key := datastore.NameKey("LogEntry", entryID, nil)

	_, err := l.db.Put(ctx, key, &entry)
	if err != nil {
		return fmt.Errorf("failed to store log entry: %w", err)
	}

	return nil
}

// storePerformanceMetric stores the performance metric in the database
func (l *Logger) storePerformanceMetric(ctx context.Context, metric PerformanceMetric) error {
	metricID := uuid.New().String()
	key := datastore.NameKey("PerformanceMetric", metricID, nil)

	_, err := l.db.Put(ctx, key, &metric)
	if err != nil {
		return fmt.Errorf("failed to store performance metric: %w", err)
	}

	return nil
}

// checkAndCreateAlert checks if an alert should be created for errors
func (l *Logger) checkAndCreateAlert(ctx context.Context, entry LogEntry) error {
	now := time.Now().UTC()

	// Reset error counts if window has passed
	if now.Sub(l.lastErrorCheck) > l.alertWindow {
		l.errorCounts = make(map[string]int)
		l.lastErrorCheck = now
	}

	// Create alert key based on error type and component
	alertKey := fmt.Sprintf("%s:%s:%s", entry.Level, entry.Service, entry.Component)
	l.errorCounts[alertKey]++

	// Check if threshold is exceeded
	if l.errorCounts[alertKey] >= l.alertThreshold {
		return l.createOrUpdateAlert(ctx, entry, l.errorCounts[alertKey])
	}

	return nil
}

// createOrUpdateAlert creates or updates an error alert
func (l *Logger) createOrUpdateAlert(ctx context.Context, entry LogEntry, errorCount int) error {
	// Check for existing unresolved alert
	query := datastore.NewQuery("ErrorAlert").
		Filter("type =", entry.Level).
		Filter("service =", entry.Service).
		Filter("component =", entry.Component).
		Filter("resolved =", false).
		Limit(1)

	var alerts []ErrorAlert
	keys, err := l.db.GetAll(ctx, query, &alerts)
	if err != nil {
		return fmt.Errorf("failed to query existing alerts: %w", err)
	}

	now := time.Now().UTC()

	if len(alerts) > 0 {
		// Update existing alert
		alert := alerts[0]
		alert.ErrorCount = errorCount
		alert.LastSeen = now
		alert.UpdatedAt = now

		key := keys[0]
		_, err = l.db.Put(ctx, key, &alert)
		if err != nil {
			return fmt.Errorf("failed to update error alert: %w", err)
		}
	} else {
		// Create new alert
		alert := ErrorAlert{
			Type:        entry.Level,
			Severity:    l.getSeverityFromLevel(entry.Level),
			Title:       fmt.Sprintf("Multiple %s errors in %s/%s", entry.Level, entry.Service, entry.Component),
			Description: fmt.Sprintf("Detected %d %s errors in %s/%s within %v", errorCount, entry.Level, entry.Service, entry.Component, l.alertWindow),
			Service:     entry.Service,
			Component:   entry.Component,
			ErrorCount:  errorCount,
			FirstSeen:   now,
			LastSeen:    now,
			Resolved:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		alertID := uuid.New().String()
		key := datastore.NameKey("ErrorAlert", alertID, nil)
		_, err = l.db.Put(ctx, key, &alert)
		if err != nil {
			return fmt.Errorf("failed to create error alert: %w", err)
		}

		// Log the alert creation
		log.Printf("ERROR_ALERT_CREATED: %s - %s", alert.Type, alert.Description)
	}

	return nil
}

// getSeverityFromLevel converts log level to alert severity
func (l *Logger) getSeverityFromLevel(level string) string {
	switch level {
	case "FATAL":
		return "CRITICAL"
	case "ERROR":
		return "HIGH"
	case "WARN":
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// getStackTrace captures the current stack trace
func (l *Logger) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// GetLogEntries retrieves log entries with filtering
func (l *Logger) GetLogEntries(ctx context.Context, filters LogFilters) ([]LogEntry, error) {
	query := datastore.NewQuery("LogEntry")

	// Apply filters
	if filters.Level != "" {
		query = query.Filter("level =", filters.Level)
	}
	if filters.Service != "" {
		query = query.Filter("service =", filters.Service)
	}
	if filters.Component != "" {
		query = query.Filter("component =", filters.Component)
	}
	if filters.UserID != "" {
		query = query.Filter("user_id =", filters.UserID)
	}
	if filters.SessionID != "" {
		query = query.Filter("session_id =", filters.SessionID)
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

	var entries []LogEntry
	keys, err := l.db.GetAll(ctx, query, &entries)
	if err != nil {
		return nil, fmt.Errorf("failed to get log entries: %w", err)
	}

	for i, key := range keys {
		entries[i].ID = key.Name
	}

	return entries, nil
}

// GetErrorAlerts retrieves error alerts
func (l *Logger) GetErrorAlerts(ctx context.Context, includeResolved bool) ([]ErrorAlert, error) {
	query := datastore.NewQuery("ErrorAlert")

	if !includeResolved {
		query = query.Filter("resolved =", false)
	}

	query = query.Order("-created_at")

	var alerts []ErrorAlert
	keys, err := l.db.GetAll(ctx, query, &alerts)
	if err != nil {
		return nil, fmt.Errorf("failed to get error alerts: %w", err)
	}

	for i, key := range keys {
		alerts[i].ID = key.Name
	}

	return alerts, nil
}

// ResolveAlert marks an error alert as resolved
func (l *Logger) ResolveAlert(ctx context.Context, alertID, resolvedBy string) error {
	key := datastore.NameKey("ErrorAlert", alertID, nil)
	var alert ErrorAlert
	err := l.db.Get(ctx, key, &alert)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alert.Resolved = true
	alert.ResolvedBy = resolvedBy
	alert.ResolvedAt = time.Now().UTC()
	alert.UpdatedAt = time.Now().UTC()

	_, err = l.db.Put(ctx, key, &alert)
	if err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	return nil
}

// LogFilters holds filters for querying log entries
type LogFilters struct {
	Level     string
	Service   string
	Component string
	UserID    string
	SessionID string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// CleanupOldLogs removes old log entries based on retention policy
func (l *Logger) CleanupOldLogs(ctx context.Context) error {
	cutoffDate := time.Now().UTC().AddDate(0, 0, -l.config.MaxRetentionDays)

	query := datastore.NewQuery("LogEntry").
		Filter("timestamp <", cutoffDate).
		KeysOnly()

	keys, err := l.db.GetAll(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query old log entries: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete in batches
	batchSize := l.config.BatchSize
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := l.db.DeleteMulti(ctx, batch); err != nil {
			return fmt.Errorf("failed to delete log entries batch: %w", err)
		}
	}

	l.Info(fmt.Sprintf("Cleaned up %d old log entries", len(keys)))
	return nil
}
