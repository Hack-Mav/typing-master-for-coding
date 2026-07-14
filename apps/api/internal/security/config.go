package security

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	Scanner    *ScannerConfig
	Logger     *LoggerConfig
	Dependency *DependencyScanConfig
	General    *GeneralSecurityConfig
}

// GeneralSecurityConfig holds general security settings
type GeneralSecurityConfig struct {
	EnableSecurityHeaders   bool
	EnableCSRFProtection    bool
	EnableRateLimiting      bool
	EnableIPWhitelist       bool
	AllowedAdminIPs         []string
	MaxRequestSize          int64
	SessionTimeout          time.Duration
	PasswordMinLength       int
	PasswordRequireSpecial  bool
	PasswordRequireNumbers  bool
	PasswordRequireUpper    bool
	PasswordRequireLower    bool
	AccountLockoutThreshold int
	AccountLockoutDuration  time.Duration
}

// LoadSecurityConfig loads security configuration from environment variables
func LoadSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Scanner:    loadScannerConfig(),
		Logger:     loadLoggerConfig(),
		Dependency: loadDependencyScanConfig(),
		General:    loadGeneralSecurityConfig(),
	}
}

// loadScannerConfig loads scanner configuration from environment
func loadScannerConfig() *ScannerConfig {
	config := DefaultScannerConfig()

	if val := os.Getenv("SECURITY_MAX_REQUEST_SIZE"); val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.MaxRequestSize = size
		}
	}

	return config
}

// loadLoggerConfig loads logger configuration from environment
func loadLoggerConfig() *LoggerConfig {
	config := DefaultLoggerConfig()

	if val := os.Getenv("SECURITY_ENABLE_FILE_LOGGING"); val != "" {
		config.EnableFileLogging = val == "true"
	}

	if val := os.Getenv("SECURITY_ENABLE_DATABASE_LOGGING"); val != "" {
		config.EnableDatabaseLogging = val == "true"
	}

	if val := os.Getenv("SECURITY_ENABLE_ALERTING"); val != "" {
		config.EnableAlerting = val == "true"
	}

	if val := os.Getenv("SECURITY_ALERT_THRESHOLD"); val != "" {
		if threshold, err := strconv.Atoi(val); err == nil {
			config.AlertThreshold = threshold
		}
	}

	if val := os.Getenv("SECURITY_ALERT_WINDOW"); val != "" {
		if window, err := time.ParseDuration(val); err == nil {
			config.AlertWindow = window
		}
	}

	if val := os.Getenv("SECURITY_LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}

	return config
}

// loadDependencyScanConfig loads dependency scan configuration from environment
func loadDependencyScanConfig() *DependencyScanConfig {
	config := DefaultDependencyScanConfig()

	if val := os.Getenv("SECURITY_ENABLE_GO_MOD_SCAN"); val != "" {
		config.EnableGoModScan = val == "true"
	}

	if val := os.Getenv("SECURITY_ENABLE_NPM_SCAN"); val != "" {
		config.EnableNpmScan = val == "true"
	}

	if val := os.Getenv("SECURITY_SCAN_INTERVAL"); val != "" {
		if interval, err := time.ParseDuration(val); err == nil {
			config.ScanInterval = interval
		}
	}

	if val := os.Getenv("SECURITY_VULNERABILITY_DB_URL"); val != "" {
		config.VulnerabilityDBURL = val
	}

	if val := os.Getenv("SECURITY_ALERT_ON_HIGH_SEVERITY"); val != "" {
		config.AlertOnHighSeverity = val == "true"
	}

	if val := os.Getenv("SECURITY_ALERT_ON_MEDIUM_SEVERITY"); val != "" {
		config.AlertOnMediumSeverity = val == "true"
	}

	return config
}

// loadGeneralSecurityConfig loads general security configuration from environment
func loadGeneralSecurityConfig() *GeneralSecurityConfig {
	config := &GeneralSecurityConfig{
		EnableSecurityHeaders:   true,
		EnableCSRFProtection:    true,
		EnableRateLimiting:      true,
		EnableIPWhitelist:       false,
		AllowedAdminIPs:         []string{},
		MaxRequestSize:          10 * 1024 * 1024, // 10MB
		SessionTimeout:          24 * time.Hour,
		PasswordMinLength:       8,
		PasswordRequireSpecial:  true,
		PasswordRequireNumbers:  true,
		PasswordRequireUpper:    true,
		PasswordRequireLower:    true,
		AccountLockoutThreshold: 5,
		AccountLockoutDuration:  30 * time.Minute,
	}

	if val := os.Getenv("SECURITY_ENABLE_SECURITY_HEADERS"); val != "" {
		config.EnableSecurityHeaders = val == "true"
	}

	if val := os.Getenv("SECURITY_ENABLE_CSRF_PROTECTION"); val != "" {
		config.EnableCSRFProtection = val == "true"
	}

	if val := os.Getenv("SECURITY_ENABLE_RATE_LIMITING"); val != "" {
		config.EnableRateLimiting = val == "true"
	}

	if val := os.Getenv("SECURITY_ENABLE_IP_WHITELIST"); val != "" {
		config.EnableIPWhitelist = val == "true"
	}

	if val := os.Getenv("SECURITY_ALLOWED_ADMIN_IPS"); val != "" {
		config.AllowedAdminIPs = strings.Split(val, ",")
	}

	if val := os.Getenv("SECURITY_MAX_REQUEST_SIZE"); val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.MaxRequestSize = size
		}
	}

	if val := os.Getenv("SECURITY_SESSION_TIMEOUT"); val != "" {
		if timeout, err := time.ParseDuration(val); err == nil {
			config.SessionTimeout = timeout
		}
	}

	if val := os.Getenv("SECURITY_PASSWORD_MIN_LENGTH"); val != "" {
		if length, err := strconv.Atoi(val); err == nil {
			config.PasswordMinLength = length
		}
	}

	if val := os.Getenv("SECURITY_PASSWORD_REQUIRE_SPECIAL"); val != "" {
		config.PasswordRequireSpecial = val == "true"
	}

	if val := os.Getenv("SECURITY_PASSWORD_REQUIRE_NUMBERS"); val != "" {
		config.PasswordRequireNumbers = val == "true"
	}

	if val := os.Getenv("SECURITY_PASSWORD_REQUIRE_UPPER"); val != "" {
		config.PasswordRequireUpper = val == "true"
	}

	if val := os.Getenv("SECURITY_PASSWORD_REQUIRE_LOWER"); val != "" {
		config.PasswordRequireLower = val == "true"
	}

	if val := os.Getenv("SECURITY_ACCOUNT_LOCKOUT_THRESHOLD"); val != "" {
		if threshold, err := strconv.Atoi(val); err == nil {
			config.AccountLockoutThreshold = threshold
		}
	}

	if val := os.Getenv("SECURITY_ACCOUNT_LOCKOUT_DURATION"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.AccountLockoutDuration = duration
		}
	}

	return config
}

// Validate validates the security configuration
func (sc *SecurityConfig) Validate() error {
	// Add validation logic here if needed
	return nil
}

// GetSecurityHeaders returns security headers configuration
func (sc *SecurityConfig) GetSecurityHeaders() map[string]string {
	headers := make(map[string]string)

	if sc.General.EnableSecurityHeaders {
		headers["X-XSS-Protection"] = "1; mode=block"
		headers["X-Content-Type-Options"] = "nosniff"
		headers["X-Frame-Options"] = "DENY"
		headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
		headers["Content-Security-Policy"] = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
		headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
		headers["Permissions-Policy"] = "geolocation=(), microphone=(), camera=()"
	}

	return headers
}
