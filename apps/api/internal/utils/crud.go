// Package utils provides utility functions for the typing master backend
package utils

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// Generic CRUD operation helpers

// GetEntityByID retrieves an entity by ID with standardized error handling
func GetEntityByID[T any](c *gin.Context, db *database.DatastoreClient, kind, id string, entity *T) error {
	key := GenerateEntityKey(kind, id)
	ctx := context.Background()

	err := db.Get(ctx, key, entity)
	if err != nil {
		HandleNotFound(c, kind)
		return err
	}

	return nil
}

// CreateEntity creates a new entity with standardized key generation and error handling
func CreateEntity[T any](c *gin.Context, db *database.DatastoreClient, kind string, entity *T, setTimestamps func(*T)) (*database.Key, error) {
	ctx := context.Background()

	// Set timestamps if function provided
	if setTimestamps != nil {
		setTimestamps(entity)
	}

	// For now, use a simple key generation - in a real implementation,
	// you'd need to determine the key structure based on the entity type
	key := database.IncompleteKey(kind, nil)

	key, err := db.Put(ctx, key, entity)
	if err != nil {
		HandleInternalError(c, err, "create "+strings.ToLower(kind))
		return nil, err
	}

	return key, nil
}

// UpdateEntity updates an existing entity with standardized error handling
func UpdateEntity[T any](c *gin.Context, db *database.DatastoreClient, kind, id string, existing *T, updates *T, applyUpdates func(*T, *T)) error {
	ctx := context.Background()

	// Apply updates
	if applyUpdates != nil {
		applyUpdates(existing, updates)
	}

	key := GenerateEntityKey(kind, id)
	_, err := db.Put(ctx, key, existing)
	if err != nil {
		HandleInternalError(c, err, "update "+strings.ToLower(kind))
		return err
	}

	return nil
}

// DeleteEntity deletes an entity by ID with standardized error handling
func DeleteEntity(c *gin.Context, db *database.DatastoreClient, kind, id string) error {
	ctx := context.Background()
	key := GenerateEntityKey(kind, id)

	err := db.Delete(ctx, key)
	if err != nil {
		HandleInternalError(c, err, "delete "+strings.ToLower(kind))
		return err
	}

	HandleSuccess(c, kind+" deleted successfully", nil)
	return nil
}

// GetEntities retrieves all entities of a kind with standardized error handling
func GetEntities[T any](c *gin.Context, db *database.DatastoreClient, kind string, entities *[]T) error {
	ctx := context.Background()
	query := database.NewQuery(kind)

	keys, err := db.GetAll(ctx, query, entities)
	if err != nil {
		HandleInternalError(c, err, "fetch "+strings.ToLower(kind)+"s")
		return err
	}

	// Set IDs for all entities
	for i, key := range keys {
		// This is a generic approach - in practice, you'd need type-specific ID handling
		// For now, we'll assume entities have an ID field that can be set
		setEntityID(entities, i, key.Name)
	}

	return nil
}

// Helper function to set entity ID (generic approach)
func setEntityID(entities interface{}, index int, id string) {
	// This is a simplified approach. In a real implementation, you'd use reflection
	// or interface{} with type assertions to set the ID field appropriately
	switch v := entities.(type) {
	case *[]models.Language:
		if index < len(*v) {
			(*v)[index].ID = id
		}
	case *[]models.Lesson:
		if index < len(*v) {
			(*v)[index].ID = id
		}
	case *[]models.Snippet:
		if index < len(*v) {
			(*v)[index].ID = id
		}
	case *[]models.Playlist:
		if index < len(*v) {
			(*v)[index].ID = id
		}
	}
}

// SetTimestamps is a generic function to set created/updated timestamps
func SetTimestamps(entity interface{}) {
	now := GetCurrentUTCTime()

	switch v := entity.(type) {
	case *models.Language:
		v.CreatedAt = now
	case *models.Lesson:
		v.CreatedAt = now
	case *models.Snippet:
		v.CreatedAt = now
	case *models.Playlist:
		v.CreatedAt = now
		v.UpdatedAt = now
	case *models.Session:
		v.CreatedAt = now
		v.StartedAt = now
	case *models.ContentVersion:
		v.CreatedAt = now
	case *models.ContentValidation:
		v.ValidatedAt = now
	}
}

// UpdateTimestamp is a generic function to update the updated timestamp
func UpdateTimestamp(entity interface{}) {
	now := GetCurrentUTCTime()

	switch v := entity.(type) {
	case *models.Playlist:
		v.UpdatedAt = now
	case *models.LessonProgress:
		v.UpdatedAt = now
	}
}
