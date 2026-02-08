package errors

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/typing-master-for-coding-backend/internal/logging"

	"github.com/gin-gonic/gin"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ValidationError     ErrorType = "VALIDATION_ERROR"
	AuthenticationError ErrorType = "AUTHENTICATION_ERROR"
	AuthorizationError  ErrorType = "AUTHORIZATION_ERROR"
	NotFoundError       ErrorType = "NOT_FOUND_ERROR"
	ConflictError       ErrorType = "CONFLICT_ERROR"
	RateLimitError      ErrorType = "RATE_LIMIT_ERROR"
	InternalError       ErrorType = "INTERNAL_ERROR"
	ServiceError        ErrorType = "SERVICE_ERROR"
	NetworkError        ErrorType = "NETWORK_ERROR"
	ParseError          ErrorType = "PARSE_ERROR"
	TimeoutError        ErrorType = "TIMEOUT_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	Type            ErrorType              `json:"type"`
	Code            string                 `json:"code"`
	Message         string                 `json:"message"`
	UserMessage     string                 `json:"user_message"`
	Details         map[string]interface{} `json:"details,omitempty"`
	Cause           error                  `json:"-"`
	StackTrace      string                 `json:"stack_trace,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	RequestID       string                 `json:"request_id,omitempty"`
	UserID          string                 `json:"user_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	Retryable       bool                   `json:"retryable"`
	RetryAfter      *time.Duration         `json:"retry_after,omitempty"`
	SuggestedAction string                 `json:"suggested_action,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger *logging.Logger
	config *ErrorHandlerConfig
}

// ErrorHandlerConfig holds configuration for error handling
type ErrorHandlerConfig struct {
	IncludeStackTrace bool          `json:"include_stack_trace"`
	LogAllErrors      bool          `json:"log_all_errors"`
	EnableRetry       bool          `json:"enable_retry"`
	DefaultRetryDelay time.Duration `json:"default_retry_delay"`
	MaxRetryAttempts  int           `json:"max_retry_attempts"`
	EnableFallbacks   bool          `json:"enable_fallbacks"`
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *logging.Logger, config *ErrorHandlerConfig) *ErrorHandler {
	if config == nil {
		config = DefaultErrorHandlerConfig()
	}

	return &ErrorHandler{
		logger: logger,
		config: config,
	}
}

// DefaultErrorHandlerConfig returns default error handler configuration
func DefaultErrorHandlerConfig() *ErrorHandlerConfig {
	return &ErrorHandlerConfig{
		IncludeStackTrace: false, // Only in development
		LogAllErrors:      true,
		EnableRetry:       true,
		DefaultRetryDelay: time.Second * 5,
		MaxRetryAttempts:  3,
		EnableFallbacks:   true,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message, userMessage string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:            ValidationError,
		Code:            "VALIDATION_FAILED",
		Message:         message,
		UserMessage:     userMessage,
		Details:         details,
		Timestamp:       time.Now().UTC(),
		Retryable:       false,
		SuggestedAction: "Please check your input and try again",
	}
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message, userMessage string) *AppError {
	return &AppError{
		Type:            AuthenticationError,
		Code:            "AUTHENTICATION_FAILED",
		Message:         message,
		UserMessage:     userMessage,
		Timestamp:       time.Now().UTC(),
		Retryable:       false,
		SuggestedAction: "Please log in and try again",
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message, userMessage string) *AppError {
	return &AppError{
		Type:            AuthorizationError,
		Code:            "AUTHORIZATION_FAILED",
		Message:         message,
		UserMessage:     userMessage,
		Timestamp:       time.Now().UTC(),
		Retryable:       false,
		SuggestedAction: "You don't have permission to perform this action",
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, userMessage string) *AppError {
	return &AppError{
		Type:        NotFoundError,
		Code:        "RESOURCE_NOT_FOUND",
		Message:     fmt.Sprintf("%s not found", resource),
		UserMessage: userMessage,
		Details: map[string]interface{}{
			"resource": resource,
		},
		Timestamp:       time.Now().UTC(),
		Retryable:       false,
		SuggestedAction: "Please check the resource identifier and try again",
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message, userMessage string, cause error) *AppError {
	err := &AppError{
		Type:            InternalError,
		Code:            "INTERNAL_SERVER_ERROR",
		Message:         message,
		UserMessage:     userMessage,
		Cause:           cause,
		Timestamp:       time.Now().UTC(),
		Retryable:       true,
		RetryAfter:      &[]time.Duration{time.Second * 5}[0],
		SuggestedAction: "Please try again in a few moments",
	}

	// Capture stack trace for internal errors
	err.StackTrace = captureStackTrace()
	return err
}

// NewServiceError creates a service error (for external service failures)
func NewServiceError(service, message, userMessage string, cause error) *AppError {
	retryDelay := time.Second * 10
	return &AppError{
		Type:        ServiceError,
		Code:        "SERVICE_UNAVAILABLE",
		Message:     fmt.Sprintf("%s service error: %s", service, message),
		UserMessage: userMessage,
		Details: map[string]interface{}{
			"service": service,
		},
		Cause:           cause,
		Timestamp:       time.Now().UTC(),
		Retryable:       true,
		RetryAfter:      &retryDelay,
		SuggestedAction: "The service is temporarily unavailable. Please try again later",
	}
}

// NewNetworkError creates a network error
func NewNetworkError(message, userMessage string, cause error) *AppError {
	retryDelay := time.Second * 3
	return &AppError{
		Type:            NetworkError,
		Code:            "NETWORK_ERROR",
		Message:         message,
		UserMessage:     userMessage,
		Cause:           cause,
		Timestamp:       time.Now().UTC(),
		Retryable:       true,
		RetryAfter:      &retryDelay,
		SuggestedAction: "Please check your connection and try again",
	}
}

// NewParseError creates a parsing error
func NewParseError(content, message, userMessage string, cause error) *AppError {
	return &AppError{
		Type:        ParseError,
		Code:        "PARSE_ERROR",
		Message:     message,
		UserMessage: userMessage,
		Details: map[string]interface{}{
			"content_type": content,
		},
		Cause:           cause,
		Timestamp:       time.Now().UTC(),
		Retryable:       false,
		SuggestedAction: "Please check the format of your input",
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation, userMessage string, timeout time.Duration) *AppError {
	retryDelay := timeout * 2
	return &AppError{
		Type:        TimeoutError,
		Code:        "OPERATION_TIMEOUT",
		Message:     fmt.Sprintf("Operation '%s' timed out after %v", operation, timeout),
		UserMessage: userMessage,
		Details: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout.String(),
		},
		Timestamp:       time.Now().UTC(),
		Retryable:       true,
		RetryAfter:      &retryDelay,
		SuggestedAction: "The operation is taking longer than expected. Please try again",
	}
}

// HandleError processes and responds to errors
func (eh *ErrorHandler) HandleError(c *gin.Context, err error) {
	var appErr *AppError

	// Convert to AppError if not already
	if !AsAppError(err, &appErr) {
		appErr = eh.convertToAppError(err)
	}

	// Add context information
	eh.enrichErrorWithContext(c, appErr)

	// Log the error
	if eh.config.LogAllErrors || appErr.Type == InternalError {
		eh.logError(c.Request.Context(), appErr)
	}

	// Prepare response
	response := eh.prepareErrorResponse(appErr)

	// Set appropriate HTTP status code
	statusCode := eh.getHTTPStatusCode(appErr.Type)

	// Add retry headers if applicable
	if appErr.Retryable && appErr.RetryAfter != nil {
		c.Header("Retry-After", fmt.Sprintf("%.0f", appErr.RetryAfter.Seconds()))
	}

	c.JSON(statusCode, response)
}

// HandlePanic recovers from panics and converts them to errors
func (eh *ErrorHandler) HandlePanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch x := r.(type) {
				case string:
					err = fmt.Errorf("panic: %s", x)
				case error:
					err = fmt.Errorf("panic: %w", x)
				default:
					err = fmt.Errorf("panic: %v", x)
				}

				appErr := NewInternalError(
					"Application panic occurred",
					"An unexpected error occurred. Please try again",
					err,
				)
				appErr.StackTrace = captureStackTrace()

				eh.HandleError(c, appErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// convertToAppError converts a regular error to AppError
func (eh *ErrorHandler) convertToAppError(err error) *AppError {
	message := err.Error()

	// Try to determine error type from message
	errorType := InternalError
	userMessage := "An unexpected error occurred"

	if strings.Contains(strings.ToLower(message), "not found") {
		errorType = NotFoundError
		userMessage = "The requested resource was not found"
	} else if strings.Contains(strings.ToLower(message), "timeout") {
		errorType = TimeoutError
		userMessage = "The operation timed out. Please try again"
	} else if strings.Contains(strings.ToLower(message), "network") {
		errorType = NetworkError
		userMessage = "A network error occurred. Please check your connection"
	} else if strings.Contains(strings.ToLower(message), "parse") || strings.Contains(strings.ToLower(message), "invalid") {
		errorType = ValidationError
		userMessage = "Invalid input provided"
	}

	return &AppError{
		Type:            errorType,
		Code:            string(errorType),
		Message:         message,
		UserMessage:     userMessage,
		Cause:           err,
		Timestamp:       time.Now().UTC(),
		Retryable:       errorType == InternalError || errorType == ServiceError || errorType == NetworkError || errorType == TimeoutError,
		SuggestedAction: "Please try again or contact support if the problem persists",
	}
}

// enrichErrorWithContext adds context information to the error
func (eh *ErrorHandler) enrichErrorWithContext(c *gin.Context, appErr *AppError) {
	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		appErr.RequestID = requestID
	}

	// Add user ID if available
	if userID := c.GetString("user_id"); userID != "" {
		appErr.UserID = userID
	}

	// Add session ID if available
	if sessionID := c.GetString("session_id"); sessionID != "" {
		appErr.SessionID = sessionID
	}

	// Add request details to error details
	if appErr.Details == nil {
		appErr.Details = make(map[string]interface{})
	}

	appErr.Details["method"] = c.Request.Method
	appErr.Details["path"] = c.Request.URL.Path
	appErr.Details["user_agent"] = c.Request.UserAgent()
	appErr.Details["remote_addr"] = c.ClientIP()
}

// logError logs the error with appropriate level
func (eh *ErrorHandler) logError(ctx context.Context, appErr *AppError) {
	metadata := map[string]interface{}{
		"error_type":       string(appErr.Type),
		"error_code":       appErr.Code,
		"request_id":       appErr.RequestID,
		"user_id":          appErr.UserID,
		"session_id":       appErr.SessionID,
		"retryable":        appErr.Retryable,
		"suggested_action": appErr.SuggestedAction,
	}

	// Add error details
	for k, v := range appErr.Details {
		metadata[k] = v
	}

	switch appErr.Type {
	case InternalError, ServiceError:
		eh.logger.Error(appErr.Message, appErr.Cause, metadata)
	case ValidationError, AuthenticationError, AuthorizationError, NotFoundError:
		eh.logger.Warn(appErr.Message, metadata)
	default:
		eh.logger.Info(appErr.Message, metadata)
	}
}

// prepareErrorResponse prepares the error response for the client
func (eh *ErrorHandler) prepareErrorResponse(appErr *AppError) map[string]interface{} {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"type":             string(appErr.Type),
			"code":             appErr.Code,
			"message":          appErr.UserMessage,
			"timestamp":        appErr.Timestamp.Format(time.RFC3339),
			"retryable":        appErr.Retryable,
			"suggested_action": appErr.SuggestedAction,
		},
	}

	// Add request ID for tracking
	if appErr.RequestID != "" {
		response["request_id"] = appErr.RequestID
	}

	// Add retry information if applicable
	if appErr.Retryable && appErr.RetryAfter != nil {
		response["error"].(map[string]interface{})["retry_after"] = appErr.RetryAfter.Seconds()
	}

	// Add stack trace in development mode
	if eh.config.IncludeStackTrace && appErr.StackTrace != "" {
		response["error"].(map[string]interface{})["stack_trace"] = appErr.StackTrace
	}

	// Add technical details in development mode
	if eh.config.IncludeStackTrace && appErr.Details != nil {
		response["error"].(map[string]interface{})["details"] = appErr.Details
	}

	return response
}

// getHTTPStatusCode maps error types to HTTP status codes
func (eh *ErrorHandler) getHTTPStatusCode(errorType ErrorType) int {
	switch errorType {
	case ValidationError:
		return http.StatusBadRequest
	case AuthenticationError:
		return http.StatusUnauthorized
	case AuthorizationError:
		return http.StatusForbidden
	case NotFoundError:
		return http.StatusNotFound
	case ConflictError:
		return http.StatusConflict
	case RateLimitError:
		return http.StatusTooManyRequests
	case TimeoutError:
		return http.StatusRequestTimeout
	case ServiceError, NetworkError:
		return http.StatusBadGateway
	case InternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// AsAppError checks if an error is an AppError and returns it
func AsAppError(err error, target **AppError) bool {
	if appErr, ok := err.(*AppError); ok {
		*target = appErr
		return true
	}
	return false
}

// captureStackTrace captures the current stack trace
func captureStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func() error

// WithRetry executes an operation with retry logic
func (eh *ErrorHandler) WithRetry(ctx context.Context, operation RetryableOperation, maxAttempts int) error {
	if !eh.config.EnableRetry {
		return operation()
	}

	if maxAttempts <= 0 {
		maxAttempts = eh.config.MaxRetryAttempts
	}

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		var appErr *AppError
		if AsAppError(err, &appErr) && !appErr.Retryable {
			return err
		}

		// Don't retry on last attempt
		if attempt == maxAttempts {
			break
		}

		// Calculate delay
		delay := eh.config.DefaultRetryDelay
		if AsAppError(err, &appErr) && appErr.RetryAfter != nil {
			delay = *appErr.RetryAfter
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}

		eh.logger.Info(fmt.Sprintf("Retrying operation (attempt %d/%d)", attempt+1, maxAttempts), map[string]interface{}{
			"attempt":      attempt + 1,
			"max_attempts": maxAttempts,
			"delay":        delay.String(),
			"error":        err.Error(),
		})
	}

	return lastErr
}

// FallbackOperation represents a fallback operation
type FallbackOperation func() error

// WithFallback executes an operation with fallback logic
func (eh *ErrorHandler) WithFallback(ctx context.Context, primary RetryableOperation, fallback FallbackOperation) error {
	if !eh.config.EnableFallbacks {
		return primary()
	}

	err := eh.WithRetry(ctx, primary, eh.config.MaxRetryAttempts)
	if err == nil {
		return nil
	}

	eh.logger.Warn("Primary operation failed, attempting fallback", map[string]interface{}{
		"primary_error": err.Error(),
	})

	fallbackErr := fallback()
	if fallbackErr == nil {
		eh.logger.Info("Fallback operation succeeded")
		return nil
	}

	eh.logger.Error("Both primary and fallback operations failed", fallbackErr, map[string]interface{}{
		"primary_error":  err.Error(),
		"fallback_error": fallbackErr.Error(),
	})

	// Return the original error, not the fallback error
	return err
}
