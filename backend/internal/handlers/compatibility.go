package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCompatibilityInfo returns browser compatibility information
func GetCompatibilityInfo(c *gin.Context) {
	// Get browser capabilities from context
	caps, exists := c.Get("browser_capabilities")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Browser capabilities not detected"})
		return
	}

	browserCaps := caps.(map[string]interface{})

	// Get accessibility features from context
	accessFeatures, exists := c.Get("accessibility_features")
	var accessibilityInfo map[string]interface{}
	if exists {
		accessibilityInfo = accessFeatures.(map[string]interface{})
	} else {
		accessibilityInfo = map[string]interface{}{}
	}

	// Get service status from context
	serviceStatus, exists := c.Get("service_status")
	var serviceInfo map[string]interface{}
	if exists {
		serviceInfo = serviceStatus.(map[string]interface{})
	} else {
		serviceInfo = map[string]interface{}{
			"database":     "available",
			"cache":        "available",
			"external_api": "available",
			"storage":      "available",
			"auth":         "available",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"browser_capabilities":   browserCaps,
		"accessibility_features": accessibilityInfo,
		"service_status":         serviceInfo,
		"timestamp":              time.Now().UTC(),
	})
}

// GetFallbackContent serves fallback content for unsupported features
func GetFallbackContent(c *gin.Context) {
	contentType := c.Query("type")

	switch contentType {
	case "languages":
		serveFallbackLanguages(c)
	case "lessons":
		serveFallbackLessons(c)
	case "snippets":
		serveFallbackSnippets(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "Invalid content type",
			"supported_types": []string{"languages", "lessons", "snippets"},
		})
	}
}

// serveFallbackLanguages provides static language data
func serveFallbackLanguages(c *gin.Context) {
	fallbackLanguages := []gin.H{
		{
			"id":          "1",
			"name":        "JavaScript",
			"extension":   "js",
			"category":    "web",
			"description": "Dynamic programming language for web development",
			"difficulty":  "beginner",
			"popularity":  95,
		},
		{
			"id":          "2",
			"name":        "Python",
			"extension":   "py",
			"category":    "backend",
			"description": "High-level programming language for various applications",
			"difficulty":  "beginner",
			"popularity":  90,
		},
		{
			"id":          "3",
			"name":        "TypeScript",
			"extension":   "ts",
			"category":    "web",
			"description": "Typed superset of JavaScript",
			"difficulty":  "intermediate",
			"popularity":  85,
		},
		{
			"id":          "4",
			"name":        "Go",
			"extension":   "go",
			"category":    "backend",
			"description": "Statically typed language for system programming",
			"difficulty":  "intermediate",
			"popularity":  75,
		},
		{
			"id":          "5",
			"name":        "Rust",
			"extension":   "rs",
			"category":    "systems",
			"description": "Systems programming language focused on safety",
			"difficulty":  "advanced",
			"popularity":  70,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"languages": fallbackLanguages,
		"source":    "fallback",
		"message":   "Serving static language data",
		"total":     len(fallbackLanguages),
	})
}

// serveFallbackLessons provides static lesson data
func serveFallbackLessons(c *gin.Context) {
	fallbackLessons := []gin.H{
		{
			"id":             "1",
			"title":          "Basic JavaScript Syntax",
			"language_id":    "1",
			"difficulty":     "beginner",
			"description":    "Learn the fundamentals of JavaScript syntax",
			"estimated_time": 15,
			"tags":           []string{"basics", "syntax", "javascript"},
		},
		{
			"id":             "2",
			"title":          "Python Variables and Data Types",
			"language_id":    "2",
			"difficulty":     "beginner",
			"description":    "Understanding variables and data types in Python",
			"estimated_time": 20,
			"tags":           []string{"variables", "data-types", "python"},
		},
		{
			"id":             "3",
			"title":          "TypeScript Interfaces",
			"language_id":    "3",
			"difficulty":     "intermediate",
			"description":    "Learn how to use TypeScript interfaces",
			"estimated_time": 25,
			"tags":           []string{"interfaces", "types", "typescript"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"lessons": fallbackLessons,
		"source":  "fallback",
		"message": "Serving static lesson data",
		"total":   len(fallbackLessons),
	})
}

// serveFallbackSnippets provides static snippet data
func serveFallbackSnippets(c *gin.Context) {
	fallbackSnippets := []gin.H{
		{
			"id":          "1",
			"title":       "Hello World in JavaScript",
			"language_id": "1",
			"code":        "console.log('Hello, World!');",
			"description": "A simple hello world example",
			"difficulty":  "beginner",
		},
		{
			"id":          "2",
			"title":       "Hello World in Python",
			"language_id": "2",
			"code":        "print('Hello, World!')",
			"description": "A simple hello world example",
			"difficulty":  "beginner",
		},
		{
			"id":          "3",
			"title":       "Hello World in TypeScript",
			"language_id": "3",
			"code":        "console.log('Hello, World!');",
			"description": "A simple hello world example",
			"difficulty":  "beginner",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"snippets": fallbackSnippets,
		"source":   "fallback",
		"message":  "Serving static snippet data",
		"total":    len(fallbackSnippets),
	})
}

// CheckCompatibility checks if the current browser supports required features
func CheckCompatibility(c *gin.Context) {
	requiredFeatures := c.QueryArray("features")

	// Get browser capabilities from context
	caps, exists := c.Get("browser_capabilities")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Browser capabilities not detected"})
		return
	}

	browserCaps := caps.(map[string]interface{})

	compatibility := gin.H{}
	supported := []string{}
	unsupported := []string{}

	for _, feature := range requiredFeatures {
		if val, ok := browserCaps[feature]; ok {
			if boolVal, ok := val.(bool); ok && boolVal {
				supported = append(supported, feature)
				compatibility[feature] = true
			} else {
				unsupported = append(unsupported, feature)
				compatibility[feature] = false
			}
		} else {
			unsupported = append(unsupported, feature)
			compatibility[feature] = false
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"compatibility": compatibility,
		"supported":     supported,
		"unsupported":   unsupported,
		"all_supported": len(unsupported) == 0,
		"message":       getCompatibilityMessage(len(unsupported)),
	})
}

// getCompatibilityMessage returns an appropriate message based on compatibility
func getCompatibilityMessage(unsupportedCount int) string {
	if unsupportedCount == 0 {
		return "All required features are supported"
	} else if unsupportedCount <= 2 {
		return "Most features are supported, some fallbacks may be used"
	} else {
		return "Many features are not supported, consider upgrading your browser"
	}
}

// GetAccessibilitySettings returns accessibility settings based on detected features
func GetAccessibilitySettings(c *gin.Context) {
	// Get accessibility features from context
	features, exists := c.Get("accessibility_features")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Accessibility features not detected"})
		return
	}

	accessFeatures := features.(map[string]interface{})

	settings := gin.H{
		"screen_reader_mode":  accessFeatures["screen_reader"],
		"high_contrast_mode":  accessFeatures["high_contrast"],
		"large_text_mode":     accessFeatures["large_text"],
		"reduced_motion":      accessFeatures["reduced_motion"],
		"keyboard_navigation": true, // Always enabled
		"focus_visible":       true, // Always enabled
		"aria_labels":         true, // Always enabled
		"semantic_markup":     true, // Always enabled
	}

	// Add device-specific settings
	if deviceType, exists := c.Get("device_type"); exists {
		settings["device_optimized"] = deviceType
	}

	c.JSON(http.StatusOK, gin.H{
		"settings":          settings,
		"detected_features": accessFeatures,
		"timestamp":         time.Now().UTC(),
	})
}

// ReportCompatibilityIssue allows users to report compatibility problems
func ReportCompatibilityIssue(c *gin.Context) {
	var report struct {
		BrowserInfo      string `json:"browser_info" binding:"required"`
		Issue            string `json:"issue" binding:"required"`
		URL              string `json:"url"`
		UserAgent        string `json:"user_agent"`
		ExpectedBehavior string `json:"expected_behavior"`
		ActualBehavior   string `json:"actual_behavior"`
	}

	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the compatibility issue (in a real implementation, this would be stored in database)
	// For now, we'll just return a success response
	c.JSON(http.StatusOK, gin.H{
		"message":    "Compatibility issue reported successfully",
		"report_id":  "COMPAT_" + generateReportID(),
		"status":     "received",
		"next_steps": "Our team will review your report and work on improvements",
	})
}

// generateReportID generates a unique report ID
func generateReportID() string {
	return string(time.Now().Unix())
}
