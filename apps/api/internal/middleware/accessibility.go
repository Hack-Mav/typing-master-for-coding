package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// AccessibilityFeatures represents detected accessibility features
type AccessibilityFeatures struct {
	ScreenReader      bool
	HighContrast      bool
	LargeText         bool
	ReducedMotion     bool
	KeyboardOnly      bool
	VoiceControl      bool
	AssistiveTech     bool
	PreferredLanguage string
}

// AccessibilityMiddleware detects accessibility preferences and assistive technologies
func AccessibilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		acceptLanguage := c.GetHeader("Accept-Language")

		features := detectAccessibilityFeatures(userAgent, acceptLanguage)

		// Store accessibility features in context
		c.Set("accessibility_features", features)

		// Set appropriate headers for accessibility
		if features.ScreenReader {
			c.Header("X-Screen-Reader", "detected")
		}

		if features.HighContrast {
			c.Header("X-High-Contrast", "enabled")
		}

		if features.ReducedMotion {
			c.Header("X-Reduced-Motion", "preferred")
		}

		c.Next()
	}
}

// detectAccessibilityFeatures analyzes user agent and headers for accessibility features
func detectAccessibilityFeatures(userAgent, acceptLanguage string) AccessibilityFeatures {
	features := AccessibilityFeatures{
		ScreenReader:      false,
		HighContrast:      false,
		LargeText:         false,
		ReducedMotion:     false,
		KeyboardOnly:      false,
		VoiceControl:      false,
		AssistiveTech:     false,
		PreferredLanguage: "en",
	}

	lowerUA := strings.ToLower(userAgent)

	// Detect screen readers
	screenReaderIndicators := []string{
		"jaws", "nvda", "window-eyes", "zoomtext", "dragon",
		"voiceover", "talkback", "chromevox", "orca",
	}

	for _, indicator := range screenReaderIndicators {
		if strings.Contains(lowerUA, indicator) {
			features.ScreenReader = true
			features.AssistiveTech = true
			break
		}
	}

	// Detect voice control
	voiceControlIndicators := []string{
		"dragon", "siri", "google assistant", "alexa", "cortana",
	}

	for _, indicator := range voiceControlIndicators {
		if strings.Contains(lowerUA, indicator) {
			features.VoiceControl = true
			features.AssistiveTech = true
			break
		}
	}

	// Detect other assistive technologies
	assistiveIndicators := []string{
		"screen reader", "magnifier", "on-screen keyboard",
		"switch access", "eye tracking", "braille",
	}

	for _, indicator := range assistiveIndicators {
		if strings.Contains(lowerUA, indicator) {
			features.AssistiveTech = true
			break
		}
	}

	// Parse preferred language
	if acceptLanguage != "" {
		// Extract first language from Accept-Language header
		langs := strings.Split(acceptLanguage, ",")
		if len(langs) > 0 {
			langParts := strings.Split(langs[0], "-")
			features.PreferredLanguage = langParts[0]
		}
	}

	return features
}

// AccessibilityHeadersMiddleware adds accessibility-specific headers
func AccessibilityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set accessibility-friendly headers
		c.Header("X-Accessibility-Supported", "true")
		c.Header("X-Keyboard-Navigation", "supported")
		c.Header("X-Screen-Reader-Support", "enhanced")

		// Add ARIA and semantic markup support indicators
		c.Header("X-ARIA-Support", "full")
		c.Header("X-Semantic-Markup", "html5")

		c.Next()
	}
}

// AccessibilityContentMiddleware modifies responses for better accessibility
func AccessibilityContentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Get accessibility features from context
		features, exists := c.Get("accessibility_features")
		if !exists {
			return
		}

		accessFeatures := features.(AccessibilityFeatures)

		// Modify response based on accessibility needs
		if accessFeatures.ScreenReader {
			// Add screen reader friendly headers
			c.Header("X-Content-Optimized", "screen-reader")
		}

		if accessFeatures.HighContrast {
			// Indicate high contrast mode support
			c.Header("X-High-Contrast-Available", "true")
		}

		if accessFeatures.ReducedMotion {
			// Indicate reduced motion support
			c.Header("X-Reduced-Motion-Support", "true")
		}
	}
}

// DeviceCompatibilityMiddleware handles device-specific accessibility
func DeviceCompatibilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")

		// Detect device type and adjust for accessibility
		lowerUA := strings.ToLower(userAgent)

		// Mobile accessibility considerations
		if strings.Contains(lowerUA, "mobile") || strings.Contains(lowerUA, "android") ||
			strings.Contains(lowerUA, "iphone") || strings.Contains(lowerUA, "ipad") {

			c.Header("X-Mobile-Optimized", "true")
			c.Header("X-Touch-Optimized", "true")

			// Check for mobile accessibility features
			if strings.Contains(lowerUA, "talkback") || strings.Contains(lowerUA, "voiceover") {
				c.Header("X-Mobile-Screen-Reader", "detected")
			}
		}

		// Tablet accessibility considerations
		if strings.Contains(lowerUA, "ipad") || strings.Contains(lowerUA, "tablet") {
			c.Header("X-Tablet-Optimized", "true")
		}

		c.Next()
	}
}
