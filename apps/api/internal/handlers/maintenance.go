package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

func GetCurrentSystemVersion(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		query := database.NewQuery("SystemVersion").Order("-ReleasedAt").Limit(20)
		var versions []models.SystemVersion
		keys, err := db.GetAll(ctx, query, &versions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system versions"})
			return
		}

		for i, key := range keys {
			versions[i].ID = key.Name
		}

		var current *models.SystemVersion
		for i := range versions {
			if versions[i].IsCurrent {
				v := versions[i]
				current = &v
				break
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"current":  current,
			"versions": versions,
		})
	}
}

func CreateSystemVersion(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var version models.SystemVersion
		if err := c.ShouldBindJSON(&version); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		now := time.Now().UTC()

		version.ID = ""
		if version.Channel == "" {
			version.Channel = "stable"
		}
		if version.ReleasedAt.IsZero() {
			version.ReleasedAt = now
		}
		version.CreatedAt = now
		version.CreatedBy = userID

		ctx := context.Background()

		if version.IsCurrent {
			q := database.NewQuery("SystemVersion").FilterField("IsCurrent", "=", true)
			var existing []models.SystemVersion
			existingKeys, err := db.GetAll(ctx, q, &existing)
			if err == nil && len(existing) > 0 {
				for i := range existing {
					existing[i].IsCurrent = false
				}
				entities := make([]interface{}, len(existing))
				for i := range existing {
					entities[i] = &existing[i]
				}
				_, _ = db.PutMulti(ctx, existingKeys, entities)
			}
		}

		key := database.IncompleteKey("SystemVersion", nil)
		storedKey, err := db.Put(ctx, key, &version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create system version"})
			return
		}

		version.ID = storedKey.Name
		c.JSON(http.StatusCreated, version)
	}
}

func GetSystemVersions(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		channel := c.Query("channel")

		query := database.NewQuery("SystemVersion").Order("-ReleasedAt")
		if channel != "" {
			query = query.FilterField("Channel", "=", channel)
		}

		var versions []models.SystemVersion
		keys, err := db.GetAll(ctx, query, &versions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system versions"})
			return
		}

		for i, key := range keys {
			versions[i].ID = key.Name
		}

		c.JSON(http.StatusOK, versions)
	}
}

func SetCurrentSystemVersion(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := context.Background()

		key := database.NameKey("SystemVersion", id, nil)
		var version models.SystemVersion
		if err := db.Get(ctx, key, &version); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "System version not found"})
			return
		}

		q := database.NewQuery("SystemVersion").Filter("IsCurrent =", true)
		var existing []models.SystemVersion
		existingKeys, err := db.GetAll(ctx, q, &existing)
		if err == nil && len(existing) > 0 {
			for i := range existing {
				existing[i].IsCurrent = false
			}
			entities := make([]interface{}, len(existing))
			for i := range existing {
				entities[i] = &existing[i]
			}
			_, _ = db.PutMulti(ctx, existingKeys, entities)
		}

		version.IsCurrent = true
		_, err = db.Put(ctx, key, &version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update system version"})
			return
		}

		version.ID = id
		c.JSON(http.StatusOK, version)
	}
}

func CreateComplianceStandard(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var standard models.ComplianceStandard
		if err := c.ShouldBindJSON(&standard); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		now := time.Now().UTC()

		standard.ID = ""
		standard.CreatedAt = now
		standard.UpdatedAt = now
		standard.CreatedBy = userID
		standard.UpdatedBy = userID
		if standard.Status == "" {
			standard.Status = "active"
		}
		if standard.EffectiveFrom.IsZero() {
			standard.EffectiveFrom = now
		}

		ctx := context.Background()
		key := database.IncompleteKey("ComplianceStandard", nil)
		storedKey, err := db.Put(ctx, key, &standard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create compliance standard"})
			return
		}

		standard.ID = storedKey.Name
		c.JSON(http.StatusCreated, standard)
	}
}

func GetComplianceStandards(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		status := c.Query("status")
		code := c.Query("code")

		query := database.NewQuery("ComplianceStandard").Order("-EffectiveFrom")
		if status != "" {
			query = query.FilterField("Status", "=", status)
		}
		if code != "" {
			query = query.FilterField("Code", "=", code)
		}

		var standards []models.ComplianceStandard
		keys, err := db.GetAll(ctx, query, &standards)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch compliance standards"})
			return
		}

		for i, key := range keys {
			standards[i].ID = key.Name
		}

		c.JSON(http.StatusOK, standards)
	}
}

func UpdateComplianceStandard(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := context.Background()

		key := database.NameKey("ComplianceStandard", id, nil)
		var existing models.ComplianceStandard
		if err := db.Get(ctx, key, &existing); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Compliance standard not found"})
			return
		}

		var updates models.ComplianceStandard
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		existing.Code = updates.Code
		existing.Name = updates.Name
		existing.Version = updates.Version
		existing.Jurisdiction = updates.Jurisdiction
		existing.Status = updates.Status
		existing.CoverageAreas = updates.CoverageAreas
		existing.DocumentationURL = updates.DocumentationURL
		existing.Notes = updates.Notes
		existing.EffectiveFrom = updates.EffectiveFrom
		existing.EffectiveTo = updates.EffectiveTo
		existing.UpdatedAt = time.Now().UTC()
		existing.UpdatedBy = c.GetString("user_id")

		_, err := db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update compliance standard"})
			return
		}

		existing.ID = id
		c.JSON(http.StatusOK, existing)
	}
}

func GetActiveComplianceStandards(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		now := time.Now().UTC()

		query := database.NewQuery("ComplianceStandard").
			FilterField("Status", "=", "active").
			FilterField("EffectiveFrom", "<=", now)

		var standards []models.ComplianceStandard
		keys, err := db.GetAll(ctx, query, &standards)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch compliance standards"})
			return
		}

		result := make([]models.ComplianceStandard, 0, len(standards))
		for i, key := range keys {
			standards[i].ID = key.Name
			if standards[i].EffectiveTo != nil && !standards[i].EffectiveTo.After(now) {
				continue
			}
			result = append(result, standards[i])
		}

		c.JSON(http.StatusOK, result)
	}
}
