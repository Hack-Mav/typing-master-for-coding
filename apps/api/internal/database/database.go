package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/typing-master-for-coding-backend/internal/config"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// DatastoreClient is the database client used by the rest of the application.
// It keeps the old name to avoid renaming the many callers across handlers/services.
type DatastoreClient struct {
	Client    *sqlx.DB
	ProjectID string
	IsMock    bool
	Mock      *MockDatastore
	store     storage
}

type storage interface {
	Put(ctx context.Context, key *Key, src interface{}) (*Key, error)
	PutMulti(ctx context.Context, keys []*Key, src interface{}) ([]*Key, error)
	Get(ctx context.Context, key *Key, dst interface{}) error
	GetAll(ctx context.Context, query *Query, dst interface{}) ([]*Key, error)
	Run(ctx context.Context, query *Query) Iterator
	Delete(ctx context.Context, key *Key) error
	DeleteMulti(ctx context.Context, keys []*Key) error
	Count(ctx context.Context, query *Query) (int, error)
}

// Initialize opens the Postgres connection when DATABASE_URL is provided,
// otherwise it returns an in-memory mock datastore for local development/testing.
func Initialize(cfg *config.Config) (*DatastoreClient, error) {
	if cfg.DatabaseURL == "" {
		return NewMockDatastoreClient(), nil
	}

	db, err := sqlx.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.DatabaseMaxOpenConns)
	db.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DatabaseConnMaxLifetimeMinutes) * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to run schema migrations: %w", err)
	}

	store := newPostgresStorage(db)
	return &DatastoreClient{
		Client:    db,
		ProjectID: cfg.ProjectID,
		IsMock:    false,
		store:     store,
	}, nil
}

// NewMockDatastoreClient returns a DatastoreClient backed by the in-memory mock store.
func NewMockDatastoreClient() *DatastoreClient {
	mock := NewMockDatastore()
	return &DatastoreClient{
		Client:    nil,
		ProjectID: "mock-project",
		IsMock:    true,
		Mock:      mock,
		store:     mock,
	}
}

func (dc *DatastoreClient) Put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.Put(ctx, key, src)
	}
	return dc.store.Put(ctx, key, src)
}

func (dc *DatastoreClient) PutMulti(ctx context.Context, keys []*Key, src interface{}) ([]*Key, error) {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.PutMulti(ctx, keys, src)
	}
	return dc.store.PutMulti(ctx, keys, src)
}

func (dc *DatastoreClient) Get(ctx context.Context, key *Key, dst interface{}) error {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.Get(ctx, key, dst)
	}
	return dc.store.Get(ctx, key, dst)
}

func (dc *DatastoreClient) GetAll(ctx context.Context, query *Query, dst interface{}) ([]*Key, error) {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.GetAll(ctx, query, dst)
	}
	return dc.store.GetAll(ctx, query, dst)
}

func (dc *DatastoreClient) Run(ctx context.Context, query *Query) Iterator {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.Run(ctx, query)
	}
	return dc.store.Run(ctx, query)
}

func (dc *DatastoreClient) Delete(ctx context.Context, key *Key) error {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.Delete(ctx, key)
	}
	return dc.store.Delete(ctx, key)
}

func (dc *DatastoreClient) DeleteMulti(ctx context.Context, keys []*Key) error {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.DeleteMulti(ctx, keys)
	}
	return dc.store.DeleteMulti(ctx, keys)
}

func (dc *DatastoreClient) Count(ctx context.Context, query *Query) (int, error) {
	if dc.IsMock && dc.Mock != nil {
		return dc.Mock.Count(ctx, query)
	}
	return dc.store.Count(ctx, query)
}

// NewQuery creates a new database query.
func (dc *DatastoreClient) NewQuery(kind string) *Query {
	return NewQuery(kind)
}

// NameKey creates a new key that uses a string name.
func (dc *DatastoreClient) NameKey(kind, name string, parent *Key) *Key {
	return NameKey(kind, name, parent)
}

// Close closes the underlying database connection.
func (dc *DatastoreClient) Close() error {
	if dc.Client != nil {
		return dc.Client.Close()
	}
	return nil
}

// Clear removes all data from the mock datastore (for testing).
func (dc *DatastoreClient) Clear() {
	if dc.IsMock && dc.Mock != nil {
		dc.Mock.Clear()
	}
}

// PutEntity stores a generic entity under the "Entity" kind (for testing).
func (dc *DatastoreClient) PutEntity(id string, entity interface{}) {
	if dc.IsMock && dc.Mock != nil {
		key := dc.NameKey("Entity", id, nil)
		dc.Mock.Put(context.Background(), key, entity)
	}
}

// PutLesson stores a lesson (for testing).
func (dc *DatastoreClient) PutLesson(id string, lesson *models.Lesson) {
	if dc.IsMock && dc.Mock != nil {
		dc.Mock.PutLesson(id, lesson)
	}
}

// PutLessonProgress stores lesson progress (for testing).
func (dc *DatastoreClient) PutLessonProgress(id string, progress *models.LessonProgress) {
	if dc.IsMock && dc.Mock != nil {
		dc.Mock.PutLessonProgress(id, progress)
	}
}

func init() {
	stdlib.GetDefaultDriver()
}
