package security

import (
	"context"
	"fmt"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
)

// AuditService handles security auditing and compliance
type AuditService struct {
	db     *database.DatastoreClient
	logger *SecurityLogger
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID         string                 `datastore:"-" json:"id"`
	Type       string                 `datastore:"type" json:"type"`
	Action     string                 `datastore:"action" json:"action"`
	Resource   string                 `datastore:"resource" json:"resource"`
	ResourceID string                 `datastore:"resource_id,omitempty" json:"resource_id,omitempty"`
	UserID     string                 `datastore:"user_id" json:"user_id"`
	UserRole   string                 `datastore:"user_role,omitempty" json:"user_role,omitempty"`
	IP         string                 `datastore:"ip" json:"ip"`
	UserAgent  string                 `datastore:"user_agent,omitempty" json:"user_agent,omitempty"`
	Success    bool                   `datastore:"success" json:"success"`
	ErrorMsg   string                 `datastore:"error_msg,omitempty" json:"error_msg,omitempty"`
	Metadata   map[string]interface{} `datastore:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp  time.Time              `datastore:"timestamp" json:"timestamp"`
	CreatedAt  time.Time              `datastore:"created_at" json:"created_at"`
}

// ComplianceReport represents a compliance audit report
type ComplianceReport struct {
	ID               string                `json:"id"`
	Type             string                `json:"type"`
	StartDate        time.Time             `json:"start_date"`
	EndDate          time.Time             `json:"end_date"`
	TotalEvents      int                   `json:"total_events"`
	SecurityEvents   int                   `json:"security_events"`
	FailedLogins     int                   `json:"failed_logins"`
	PrivilegeChanges int                   `json:"privilege_changes"`
	DataAccess       int                   `json:"data_access"`
	Violations       []ComplianceViolation `json:"violations"`
	Recommendations  []string              `json:"recommendations"`
	ComplianceScore  float64               `json:"compliance_score"`
	GeneratedAt      time.Time             `json:"generated_at"`
	GeneratedBy      string                `json:"generated_by"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Count       int       `json:"count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

// NewAuditService creates a new audit service
func NewAuditService(db *database.DatastoreClient, logger *SecurityLogger) *AuditService {
	return &AuditService{
		db:     db,
		logger: logger,
	}
}

// LogAuditEvent logs an audit event
func (as *AuditService) LogAuditEvent(ctx context.Context, event AuditEvent) error {
	event.ID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.CreatedAt = time.Now().UTC()

	key := datastore.NameKey("AuditEvent", event.ID, nil)
	_, err := as.db.Put(ctx, key, &event)
	if err != nil {
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil
}

// GetAuditEvents retrieves audit events with filtering
func (as *AuditService) GetAuditEvents(ctx context.Context, filters AuditEventFilters) ([]AuditEvent, error) {
	query := datastore.NewQuery("AuditEvent")

	// Apply filters
	if filters.Type != "" {
		query = query.Filter("type =", filters.Type)
	}
	if filters.Action != "" {
		query = query.Filter("action =", filters.Action)
	}
	if filters.Resource != "" {
		query = query.Filter("resource =", filters.Resource)
	}
	if filters.UserID != "" {
		query = query.Filter("user_id =", filters.UserID)
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

	var events []AuditEvent
	keys, err := as.db.GetAll(ctx, query, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit events: %w", err)
	}

	for i, key := range keys {
		events[i].ID = key.Name
	}

	return events, nil
}

// GenerateComplianceReport generates a compliance audit report
func (as *AuditService) GenerateComplianceReport(ctx context.Context, startDate, endDate time.Time, reportType, generatedBy string) (*ComplianceReport, error) {
	// Get audit events in the date range
	filters := AuditEventFilters{
		StartTime: startDate,
		EndTime:   endDate,
		Limit:     10000, // Large limit for comprehensive report
	}

	auditEvents, err := as.GetAuditEvents(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit events: %w", err)
	}

	// Get security events in the same range
	securityFilters := SecurityEventFilters{
		StartTime: startDate,
		EndTime:   endDate,
		Limit:     10000,
	}

	securityEvents, err := as.logger.GetSecurityEvents(ctx, securityFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	// Analyze events and generate report
	report := &ComplianceReport{
		ID:             uuid.New().String(),
		Type:           reportType,
		StartDate:      startDate,
		EndDate:        endDate,
		TotalEvents:    len(auditEvents),
		SecurityEvents: len(securityEvents),
		GeneratedAt:    time.Now().UTC(),
		GeneratedBy:    generatedBy,
	}

	// Analyze audit events
	failedLogins := 0
	privilegeChanges := 0
	dataAccess := 0

	for _, event := range auditEvents {
		switch event.Action {
		case "login":
			if !event.Success {
				failedLogins++
			}
		case "role_assign", "role_revoke", "permission_grant", "permission_revoke":
			privilegeChanges++
		case "read", "export":
			if event.Resource == "user_data" || event.Resource == "sessions" {
				dataAccess++
			}
		}
	}

	report.FailedLogins = failedLogins
	report.PrivilegeChanges = privilegeChanges
	report.DataAccess = dataAccess

	// Identify compliance violations
	violations := as.identifyViolations(auditEvents, securityEvents)
	report.Violations = violations

	// Generate recommendations
	recommendations := as.generateRecommendations(violations, report)
	report.Recommendations = recommendations

	// Calculate compliance score
	report.ComplianceScore = as.calculateComplianceScore(report)

	return report, nil
}

// identifyViolations identifies compliance violations from events
func (as *AuditService) identifyViolations(auditEvents []AuditEvent, securityEvents []SecurityEvent) []ComplianceViolation {
	violations := []ComplianceViolation{}

	// Check for excessive failed logins
	failedLoginsByIP := make(map[string]int)
	for _, event := range auditEvents {
		if event.Action == "login" && !event.Success {
			failedLoginsByIP[event.IP]++
		}
	}

	for ip, count := range failedLoginsByIP {
		if count > 10 { // Threshold for excessive failed logins
			violations = append(violations, ComplianceViolation{
				Type:        "EXCESSIVE_FAILED_LOGINS",
				Severity:    "HIGH",
				Description: fmt.Sprintf("IP %s had %d failed login attempts", ip, count),
				Count:       count,
			})
		}
	}

	// Check for security events
	securityEventsByType := make(map[string]int)
	for _, event := range securityEvents {
		securityEventsByType[event.Type]++
	}

	for eventType, count := range securityEventsByType {
		if count > 5 { // Threshold for security violations
			violations = append(violations, ComplianceViolation{
				Type:        "SECURITY_VIOLATIONS",
				Severity:    "CRITICAL",
				Description: fmt.Sprintf("Multiple %s attempts detected (%d occurrences)", eventType, count),
				Count:       count,
			})
		}
	}

	// Check for privilege escalation patterns
	privilegeChanges := make(map[string]int)
	for _, event := range auditEvents {
		if event.Action == "role_assign" && event.UserRole == "admin" {
			privilegeChanges[event.UserID]++
		}
	}

	for userID, count := range privilegeChanges {
		if count > 1 { // Multiple admin role assignments
			violations = append(violations, ComplianceViolation{
				Type:        "PRIVILEGE_ESCALATION",
				Severity:    "HIGH",
				Description: fmt.Sprintf("User %s received admin privileges %d times", userID, count),
				Count:       count,
			})
		}
	}

	return violations
}

// generateRecommendations generates security recommendations based on violations
func (as *AuditService) generateRecommendations(violations []ComplianceViolation, report *ComplianceReport) []string {
	recommendations := []string{}

	// Check violation types and generate recommendations
	hasFailedLogins := false
	hasSecurityViolations := false
	hasPrivilegeEscalation := false

	for _, violation := range violations {
		switch violation.Type {
		case "EXCESSIVE_FAILED_LOGINS":
			hasFailedLogins = true
		case "SECURITY_VIOLATIONS":
			hasSecurityViolations = true
		case "PRIVILEGE_ESCALATION":
			hasPrivilegeEscalation = true
		}
	}

	if hasFailedLogins {
		recommendations = append(recommendations, "Implement account lockout policies after multiple failed login attempts")
		recommendations = append(recommendations, "Consider implementing CAPTCHA for login attempts from suspicious IPs")
	}

	if hasSecurityViolations {
		recommendations = append(recommendations, "Review and strengthen input validation and sanitization")
		recommendations = append(recommendations, "Implement rate limiting for API endpoints")
		recommendations = append(recommendations, "Consider implementing Web Application Firewall (WAF)")
	}

	if hasPrivilegeEscalation {
		recommendations = append(recommendations, "Implement approval workflow for privilege escalation")
		recommendations = append(recommendations, "Regular review of user roles and permissions")
		recommendations = append(recommendations, "Implement time-limited role assignments")
	}

	// General recommendations based on metrics
	if report.FailedLogins > 100 {
		recommendations = append(recommendations, "High number of failed logins detected - review authentication security")
	}

	if report.SecurityEvents > 50 {
		recommendations = append(recommendations, "High number of security events - consider strengthening security controls")
	}

	return recommendations
}

// calculateComplianceScore calculates a compliance score based on the report
func (as *AuditService) calculateComplianceScore(report *ComplianceReport) float64 {
	baseScore := 100.0

	// Deduct points for violations
	for _, violation := range report.Violations {
		switch violation.Severity {
		case "CRITICAL":
			baseScore -= 20.0
		case "HIGH":
			baseScore -= 10.0
		case "MEDIUM":
			baseScore -= 5.0
		case "LOW":
			baseScore -= 2.0
		}
	}

	// Deduct points for high numbers of security events
	if report.SecurityEvents > 100 {
		baseScore -= 10.0
	} else if report.SecurityEvents > 50 {
		baseScore -= 5.0
	}

	// Deduct points for excessive failed logins
	if report.FailedLogins > 200 {
		baseScore -= 10.0
	} else if report.FailedLogins > 100 {
		baseScore -= 5.0
	}

	// Ensure score doesn't go below 0
	if baseScore < 0 {
		baseScore = 0
	}

	return baseScore
}

// AuditEventFilters holds filters for querying audit events
type AuditEventFilters struct {
	Type      string
	Action    string
	Resource  string
	UserID    string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}
