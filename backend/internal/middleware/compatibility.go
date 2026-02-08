package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// BrowserCapabilities represents detected browser features
type BrowserCapabilities struct {
	SupportsWebP          bool
	SupportsAVIF          bool
	SupportsES6           bool
	SupportsWASM          bool
	SupportsServiceWorker bool
	SupportsPushAPI       bool
	IsMobile              bool
	IsTouchDevice         bool
	PreferredFormat       string
}

// FeatureDetectionMiddleware detects browser capabilities and sets appropriate headers
func FeatureDetectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		acceptHeader := c.GetHeader("Accept")

		capabilities := detectBrowserCapabilities(userAgent, acceptHeader)

		// Store capabilities in context for use in handlers
		c.Set("browser_capabilities", capabilities)

		// Set appropriate response headers based on capabilities
		if capabilities.SupportsWebP {
			c.Header("Vary", "Accept")
		}

		if capabilities.IsMobile {
			c.Header("X-Device-Type", "mobile")
		} else {
			c.Header("X-Device-Type", "desktop")
		}

		c.Next()
	}
}

// detectBrowserCapabilities analyzes user agent and accept headers
func detectBrowserCapabilities(userAgent, acceptHeader string) BrowserCapabilities {
	caps := BrowserCapabilities{
		SupportsWebP:          false,
		SupportsAVIF:          false,
		SupportsES6:           false,
		SupportsWASM:          false,
		SupportsServiceWorker: false,
		SupportsPushAPI:       false,
		IsMobile:              false,
		IsTouchDevice:         false,
		PreferredFormat:       "png",
	}

	// Detect mobile devices
	mobileRegex := regexp.MustCompile(`(?i)(android|iphone|ipad|ipod|blackberry|windows phone|mobile|opera mini|iemobile)`)
	caps.IsMobile = mobileRegex.MatchString(userAgent)

	// Detect touch devices
	touchRegex := regexp.MustCompile(`(?i)(touch|tablet|ipad|android|iphone|ipod)`)
	caps.IsTouchDevice = touchRegex.MatchString(userAgent)

	// Detect WebP support
	if strings.Contains(acceptHeader, "image/webp") {
		caps.SupportsWebP = true
		caps.PreferredFormat = "webp"
	}

	// Detect AVIF support
	if strings.Contains(acceptHeader, "image/avif") {
		caps.SupportsAVIF = true
		caps.PreferredFormat = "avif"
	}

	// Detect browser versions for ES6, WASM, Service Worker support
	lowerUA := strings.ToLower(userAgent)

	// Chrome 51+ supports ES6, WASM, Service Worker
	if strings.Contains(lowerUA, "chrome/") {
		if version := extractVersion(lowerUA, "chrome/"); version >= 51 {
			caps.SupportsES6 = true
			caps.SupportsWASM = true
			caps.SupportsServiceWorker = true
		}
	}

	// Firefox 45+ supports ES6, Service Worker; 52+ supports WASM
	if strings.Contains(lowerUA, "firefox/") {
		if version := extractVersion(lowerUA, "firefox/"); version >= 45 {
			caps.SupportsES6 = true
			caps.SupportsServiceWorker = true
		}
		if version := extractVersion(lowerUA, "firefox/"); version >= 52 {
			caps.SupportsWASM = true
		}
	}

	// Safari 10.1+ supports ES6, Service Worker
	if strings.Contains(lowerUA, "safari/") && !strings.Contains(lowerUA, "chrome") {
		if version := extractVersion(lowerUA, "version/"); version >= 10 {
			caps.SupportsES6 = true
			caps.SupportsServiceWorker = true
		}
	}

	// Edge 16+ supports ES6, WASM, Service Worker, Push API
	if strings.Contains(lowerUA, "edg/") {
		if version := extractVersion(lowerUA, "edg/"); version >= 16 {
			caps.SupportsES6 = true
			caps.SupportsWASM = true
			caps.SupportsServiceWorker = true
			caps.SupportsPushAPI = true
		}
	}

	return caps
}

// extractVersion extracts browser version from user agent string
func extractVersion(userAgent, prefix string) int {
	re := regexp.MustCompile(prefix + `(\d+)`)
	matches := re.FindStringSubmatch(userAgent)
	if len(matches) > 1 {
		var version int
		_, err := fmt.Sscanf(matches[1], "%d", &version)
		if err == nil {
			return version
		}
	}
	return 0
}

// CompatibilityHeadersMiddleware adds headers for cross-platform compatibility
func CompatibilityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set security headers for broad compatibility
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Set caching headers for better performance
		if c.Request.Method == "GET" {
			c.Header("Cache-Control", "public, max-age=3600")
		}

		// Set compatibility mode for IE
		c.Header("X-UA-Compatible", "IE=edge")

		c.Next()
	}
}

// FallbackMiddleware provides fallback responses for unsupported features
func FallbackMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get browser capabilities from context
		caps, exists := c.Get("browser_capabilities")
		if !exists {
			caps = BrowserCapabilities{} // Default to minimal capabilities
		}

		browserCaps := caps.(BrowserCapabilities)

		// Check for unsupported features and provide fallbacks
		path := c.Request.URL.Path

		// Fallback for modern API endpoints
		if strings.Contains(path, "/api/v2/") && !browserCaps.SupportsES6 {
			c.JSON(http.StatusNotFound, gin.H{
				"error":    "API version not supported",
				"fallback": "/api/v1/",
				"message":  "Please use API v1 for better browser compatibility",
			})
			c.Abort()
			return
		}

		// Fallback for WASM-based features
		if strings.Contains(path, "/wasm/") && !browserCaps.SupportsWASM {
			c.JSON(http.StatusNotFound, gin.H{
				"error":    "WebAssembly not supported",
				"fallback": "/js/",
				"message":  "JavaScript fallback available",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
