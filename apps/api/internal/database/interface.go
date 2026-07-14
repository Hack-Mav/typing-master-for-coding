package database

import "context"

// DatastoreInterface defines the interface for database operations.
type DatastoreInterface interface {
	Put(ctx context.Context, key *Key, src interface{}) (*Key, error)
	PutMulti(ctx context.Context, keys []*Key, src interface{}) ([]*Key, error)
	Get(ctx context.Context, key *Key, dst interface{}) error
	GetAll(ctx context.Context, q *Query, dst interface{}) ([]*Key, error)
	Run(ctx context.Context, q *Query) Iterator
	Delete(ctx context.Context, key *Key) error
	DeleteMulti(ctx context.Context, keys []*Key) error
	Count(ctx context.Context, q *Query) (int, error)
	NewQuery(kind string) *Query
	NameKey(kind, name string, parent *Key) *Key
	Close() error
}

// Ensure DatastoreClient implements DatastoreInterface.
var _ DatastoreInterface = (*DatastoreClient)(nil)
