package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"typing-master-backend/internal/telemetry"
)

// TelemetryMiddleware adds OpenTelemetry tracing and metrics to requests
func TelemetryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment active connections
		telemetry.IncrementActiveConnections(c.Request.Context())
		defer telemetry.DecrementActiveConnections(c.Request.Context())

		// Start span
		ctx, span := telemetry.StartSpan(c.Request.Context(), c.Request.URL.Path)
		defer span.End()

		// Update context
		c.Request = c.Request.WithContext(ctx)

		// Record start time
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Record metrics
		telemetry.RecordRequest(
			ctx,
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			duration,
		)

		// Record errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				// Convert error type to string
				errorType := fmt.Sprintf("%d", err.Type)
				telemetry.RecordError(ctx, errorType, c.FullPath())
			}
		}
	}
}
