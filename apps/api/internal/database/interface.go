package database

import (
	"context"

	"cloud.google.com/go/datastore"
)

// DatastoreInterface defines the interface for datastore operations
// This allows other packages to use database operations without importing the concrete implementation
type DatastoreInterface interface {
	Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error)
	PutMulti(ctx context.Context, keys []*datastore.Key, src interface{}) ([]*datastore.Key, error)
	Get(ctx context.Context, key *datastore.Key, dst interface{}) error
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error)
	NewQuery(kind string) *datastore.Query
	NameKey(kind, name string, parent *datastore.Key) *datastore.Key
	Delete(ctx context.Context, key *datastore.Key) error
	DeleteMulti(ctx context.Context, keys []*datastore.Key) error
	Count(ctx context.Context, q *datastore.Query) (int, error)
}

// Ensure DatastoreClient implements the interface
var _ DatastoreInterface = (*DatastoreClient)(nil)
