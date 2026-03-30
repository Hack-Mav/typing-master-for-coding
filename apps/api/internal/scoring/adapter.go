package scoring

import (
	"context"

	"cloud.google.com/go/datastore"
)

// datastoreClientAdapter wraps a *datastore.Client to implement the DatastoreClient interface
type datastoreClientAdapter struct {
	client *datastore.Client
}

// NewDatastoreClient creates a new DatastoreClient that wraps a *datastore.Client
func NewDatastoreClient(client *datastore.Client) DatastoreClient {
	return &datastoreClientAdapter{client: client}
}

func (a *datastoreClientAdapter) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	return a.client.Put(ctx, key, src)
}

func (a *datastoreClientAdapter) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	return a.client.Get(ctx, key, dst)
}

func (a *datastoreClientAdapter) Run(ctx context.Context, q *datastore.Query) Iterator {
	return a.client.Run(ctx, q)
}

func (a *datastoreClientAdapter) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	return a.client.GetAll(ctx, q, dst)
}
