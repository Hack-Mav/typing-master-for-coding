// Package utils provides utility functions for the typing master backend
package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Content processing utilities

// GenerateChecksum creates a SHA256 checksum for the given content
func GenerateChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// NormalizeCode performs basic normalization on code content
func NormalizeCode(code string) string {
	// Basic normalization: trim whitespace, normalize line endings
	lines := strings.Split(strings.TrimSpace(code), "\n")
	var normalized []string
	for _, line := range lines {
		normalized = append(normalized, strings.TrimSpace(line))
	}
	return strings.Join(normalized, "\n")
}

// AssessCodeDifficulty determines code difficulty based on line count and complexity indicators
func AssessCodeDifficulty(code string) int {
	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	// Simple heuristic based on line count and complexity indicators
	if lineCount <= 10 {
		return 1 // Easy
	} else if lineCount <= 30 {
		return 2 // Medium
	} else if lineCount <= 60 {
		return 3 // Hard
	} else {
		return 4 // Expert
	}
}

// GenerateTagsFromCode automatically generates tags based on code patterns
func GenerateTagsFromCode(code string) []string {
	var tags []string

	// Simple tag generation based on code patterns
	if strings.Contains(code, "function") || strings.Contains(code, "def ") {
		tags = append(tags, "function")
	}
	if strings.Contains(code, "class") {
		tags = append(tags, "class")
	}
	if strings.Contains(code, "if ") || strings.Contains(code, "if(") {
		tags = append(tags, "conditional")
	}
	if strings.Contains(code, "for ") || strings.Contains(code, "for(") {
		tags = append(tags, "loop")
	}
	if strings.Contains(code, "import") || strings.Contains(code, "from ") {
		tags = append(tags, "import")
	}

	// Default tag if no patterns found
	if len(tags) == 0 {
		tags = append(tags, "general")
	}

	return tags
}

// EstimateCodingTime estimates the time needed to complete a coding exercise
func EstimateCodingTime(code string) int {
	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	// Simple estimation: ~1 minute per 10 lines, minimum 1 minute
	estimatedMinutes := lineCount / 10
	if estimatedMinutes < 1 {
		estimatedMinutes = 1
	}

	return estimatedMinutes
}

// GenerateAccessibilityTags creates accessibility metadata for code snippets
func GenerateAccessibilityTags(code string) map[string]interface{} {
	accessibility := make(map[string]interface{})

	// Basic accessibility assessment
	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	accessibility["line_count"] = lineCount
	accessibility["estimated_time"] = EstimateCodingTime(code)
	accessibility["complexity_score"] = AssessCodeDifficulty(code)

	// Check for potential accessibility issues
	if lineCount > 50 {
		accessibility["long_content"] = true
	}

	return accessibility
}
