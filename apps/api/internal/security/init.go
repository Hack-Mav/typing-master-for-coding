package security

import (
	"context"
	"log"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
)

// SecurityManager manages all security components
type SecurityManager struct {
	Config            *SecurityConfig
	Logger            *SecurityLogger
	Scanner           *VulnerabilityScanner
	AuditService      *AuditService
	DependencyScanner *DependencyScanner
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(db *database.DatastoreClient) *SecurityManager {
	config := LoadSecurityConfig()

	logger := NewSecurityLogger(db, config.Logger)
	scanner := NewVulnerabilityScanner(config.Scanner, logger)
	auditService := NewAuditService(db, logger)
	dependencyScanner := NewDependencyScanner(config.Dependency)

	return &SecurityManager{
		Config:            config,
		Logger:            logger,
		Scanner:           scanner,
		AuditService:      auditService,
		DependencyScanner: dependencyScanner,
	}
}

// Initialize initializes all security components
func (sm *SecurityManager) Initialize(ctx context.Context) error {
	log.Println("Initializing security components...")

	// Validate configuration
	if err := sm.Config.Validate(); err != nil {
		return err
	}

	// Initialize security logging
	log.Println("Security logging initialized")

	// Log security initialization event
	initEvent := SecurityThreat{
		Type:        "SYSTEM_INIT",
		Severity:    "INFO",
		Description: "Security system initialized",
		Source:      "security_manager",
		Timestamp:   time.Now(),
		Blocked:     false,
	}

	if err := sm.Logger.LogSecurityEvent(ctx, initEvent); err != nil {
		log.Printf("Failed to log security initialization: %v", err)
	}

	log.Println("Security components initialized successfully")
	return nil
}

// StartPeriodicScans starts periodic security scans
func (sm *SecurityManager) StartPeriodicScans(ctx context.Context, projectPath string) {
	if !sm.Config.Dependency.EnableGoModScan && !sm.Config.Dependency.EnableNpmScan {
		log.Println("Dependency scanning disabled")
		return
	}

	log.Printf("Starting periodic dependency scans every %v", sm.Config.Dependency.ScanInterval)

	ticker := time.NewTicker(sm.Config.Dependency.ScanInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping periodic security scans")
				return
			case <-ticker.C:
				sm.performPeriodicScan(ctx, projectPath)
			}
		}
	}()
}

// performPeriodicScan performs a periodic dependency scan
func (sm *SecurityManager) performPeriodicScan(ctx context.Context, projectPath string) {
	log.Println("Starting periodic dependency scan...")

	result, err := sm.DependencyScanner.ScanDependencies(ctx, projectPath)
	if err != nil {
		log.Printf("Periodic dependency scan failed: %v", err)

		// Log scan failure as security event
		failureEvent := SecurityThreat{
			Type:        "SCAN_FAILURE",
			Severity:    "MEDIUM",
			Description: "Periodic dependency scan failed: " + err.Error(),
			Source:      "dependency_scanner",
			Timestamp:   time.Now(),
			Blocked:     false,
		}
		sm.Logger.LogSecurityEvent(ctx, failureEvent)
		return
	}

	log.Printf("Periodic dependency scan completed: %d vulnerabilities found", len(result.Vulnerabilities))

	// Check for high-severity vulnerabilities and create alerts
	highSeverityCount := 0
	for _, vuln := range result.Vulnerabilities {
		if vuln.Severity == "HIGH" || vuln.Severity == "CRITICAL" {
			highSeverityCount++

			if sm.Config.Dependency.AlertOnHighSeverity {
				// Log as security threat
				threat := SecurityThreat{
					Type:        "VULNERABILITY_DETECTED",
					Severity:    vuln.Severity,
					Description: "High-severity vulnerability detected: " + vuln.Title,
					Source:      "dependency_scanner",
					Timestamp:   time.Now(),
					Blocked:     false,
				}
				sm.Logger.LogSecurityEvent(ctx, threat)
			}
		}
	}

	if highSeverityCount > 0 {
		log.Printf("WARNING: %d high-severity vulnerabilities detected in dependencies", highSeverityCount)
	}
}

// GetSecurityStatus returns current security status
func (sm *SecurityManager) GetSecurityStatus(ctx context.Context) (*SecurityStatus, error) {
	// Get recent security metrics
	metrics, err := sm.Logger.GetSecurityMetrics(ctx, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	// Get active alerts
	alerts, err := sm.Logger.GetSecurityAlerts(ctx, false)
	if err != nil {
		return nil, err
	}

	// Calculate overall security score
	score := calculateOverallSecurityScore(metrics, alerts)

	status := &SecurityStatus{
		OverallScore:     score,
		TotalEvents:      metrics.TotalEvents,
		BlockedEvents:    metrics.BlockedEvents,
		ActiveAlerts:     len(alerts),
		EventsByType:     metrics.EventsByType,
		EventsBySeverity: metrics.EventsBySeverity,
		LastUpdated:      time.Now(),
		SystemStatus:     "operational",
	}

	// Determine system status based on score and alerts
	if score < 50 || len(alerts) > 10 {
		status.SystemStatus = "degraded"
	}
	if score < 30 || len(alerts) > 20 {
		status.SystemStatus = "critical"
	}

	return status, nil
}

// SecurityStatus represents the overall security status
type SecurityStatus struct {
	OverallScore     float64        `json:"overall_score"`
	TotalEvents      int            `json:"total_events"`
	BlockedEvents    int            `json:"blocked_events"`
	ActiveAlerts     int            `json:"active_alerts"`
	EventsByType     map[string]int `json:"events_by_type"`
	EventsBySeverity map[string]int `json:"events_by_severity"`
	LastUpdated      time.Time      `json:"last_updated"`
	SystemStatus     string         `json:"system_status"`
}

// calculateOverallSecurityScore calculates an overall security score
func calculateOverallSecurityScore(metrics *SecurityMetrics, alerts []SecurityAlert) float64 {
	baseScore := 100.0

	// Deduct points for security events
	if metrics.TotalEvents > 200 {
		baseScore -= 30.0
	} else if metrics.TotalEvents > 100 {
		baseScore -= 20.0
	} else if metrics.TotalEvents > 50 {
		baseScore -= 10.0
	}

	// Deduct points for active alerts
	for _, alert := range alerts {
		switch alert.Severity {
		case "CRITICAL":
			baseScore -= 20.0
		case "HIGH":
			baseScore -= 15.0
		case "MEDIUM":
			baseScore -= 10.0
		case "LOW":
			baseScore -= 5.0
		}
	}

	// Bonus for having blocked events (shows security is working)
	if metrics.BlockedEvents > 0 && metrics.TotalEvents > 0 {
		blockRate := float64(metrics.BlockedEvents) / float64(metrics.TotalEvents)
		if blockRate > 0.9 {
			baseScore += 10.0
		} else if blockRate > 0.7 {
			baseScore += 5.0
		}
	}

	// Ensure score is between 0 and 100
	if baseScore < 0 {
		baseScore = 0
	} else if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}

// Shutdown gracefully shuts down security components
func (sm *SecurityManager) Shutdown(ctx context.Context) error {
	log.Println("Shutting down security components...")

	// Log shutdown event
	shutdownEvent := SecurityThreat{
		Type:        "SYSTEM_SHUTDOWN",
		Severity:    "INFO",
		Description: "Security system shutting down",
		Source:      "security_manager",
		Timestamp:   time.Now(),
		Blocked:     false,
	}

	if err := sm.Logger.LogSecurityEvent(ctx, shutdownEvent); err != nil {
		log.Printf("Failed to log security shutdown: %v", err)
	}

	log.Println("Security components shut down successfully")
	return nil
}
