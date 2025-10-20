// Package utils provides utility functions for the typing master backend
package utils

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
)

// Key generation helpers for consistent datastore key creation

// GenerateEntityKey creates a datastore key for an existing entity by ID
func GenerateEntityKey(kind, id string) *datastore.Key {
	return datastore.NameKey(kind, id, nil)
}

// GenerateNewEntityKey creates a datastore key for a new entity with a composite ID
// Format: {parentID}_{timestamp}
func GenerateNewEntityKey(kind, parentID string) *datastore.Key {
	return datastore.NameKey(kind, fmt.Sprintf("%s_%d", parentID, time.Now().Unix()), nil)
}

// GenerateSessionKey creates a datastore key for a session
// Format: session_{userID}_{timestamp}
func GenerateSessionKey(userID string) *datastore.Key {
	return datastore.NameKey("Session", fmt.Sprintf("session_%s_%d", userID, time.Now().UnixNano()), nil)
}

// GenerateEventKey creates a datastore key for a session event
// Format: {sessionID}_event_{timestamp}_{index}
func GenerateEventKey(sessionID string, index int64) *datastore.Key {
	return datastore.NameKey("SessionEvent", fmt.Sprintf("%s_event_%d_%d", sessionID, time.Now().UnixNano(), index), nil)
}

// GenerateContentVersionKey creates a datastore key for a content version
// Format: {contentType}_{contentID}_{version}
func GenerateContentVersionKey(contentType, contentID string, version int) *datastore.Key {
	return datastore.NameKey("ContentVersion", fmt.Sprintf("%s_%s_%d", contentType, contentID, version), nil)
}

// GenerateValidationKey creates a datastore key for content validation
// Format: {contentType}_{contentID}_{timestamp}
func GenerateValidationKey(contentType, contentID string) *datastore.Key {
	return datastore.NameKey("ContentValidation", fmt.Sprintf("%s_%s_%d", contentType, contentID, time.Now().Unix()), nil)
}

// GetCurrentUTCTime returns the current time in UTC
func GetCurrentUTCTime() time.Time {
	return time.Now().UTC()
}

// GetCurrentTimestamp returns the current Unix timestamp
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
