package database

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

type DatastoreClient struct {
	Client    *datastore.Client
	ProjectID string
	IsMock    bool
	Mock      *MockDatastore
}

func Initialize(projectID string) (*DatastoreClient, error) {
	ctx := context.Background()

	// Check if we're in development mode without credentials
	environment := os.Getenv("ENVIRONMENT")
	if environment == "development" {
		// Try to use emulator first
		if emulatorHost := os.Getenv("DATASTORE_EMULATOR_HOST"); emulatorHost != "" {
			client, err := datastore.NewClient(ctx, projectID)
			if err != nil {
				return nil, fmt.Errorf("failed to create datastore emulator client: %v", err)
			}
			return &DatastoreClient{
				Client:    client,
				ProjectID: projectID,
				IsMock:    false,
				Mock:      nil,
			}, nil
		}

		// If no emulator and no credentials, return a mock client for development
		return &DatastoreClient{
			Client:    nil,
			ProjectID: projectID,
			IsMock:    true,
			Mock:      NewMockDatastore(),
		}, nil
	}

	// Production mode - use real credentials
	var client *datastore.Client
	var err error

	if credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credentialsFile != "" {
		client, err = datastore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsFile))
	} else {
		client, err = datastore.NewClient(ctx, projectID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create datastore client: %v", err)
	}

	return &DatastoreClient{
		Client:    client,
		ProjectID: projectID,
		IsMock:    false,
		Mock:      nil,
	}, nil
}

func (dc *DatastoreClient) Close() error {
	if dc.IsMock || dc.Client == nil {
		return nil
	}
	return dc.Client.Close()
}

// Helper methods that work with both real and mock datastores
func (dc *DatastoreClient) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	if dc.IsMock {
		return dc.Mock.Put(ctx, key, src)
	}
	return dc.Client.Put(ctx, key, src)
}

func (dc *DatastoreClient) PutMulti(ctx context.Context, keys []*datastore.Key, src interface{}) ([]*datastore.Key, error) {
	if dc.IsMock {
		// For mock, call Put for each item
		resultKeys := make([]*datastore.Key, len(keys))
		for i, key := range keys {
			// Extract individual item from slice
			// This is a simplified implementation for mock
			resultKey, err := dc.Mock.Put(ctx, key, src)
			if err != nil {
				return nil, err
			}
			resultKeys[i] = resultKey
		}
		return resultKeys, nil
	}
	return dc.Client.PutMulti(ctx, keys, src)
}

func (dc *DatastoreClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	if dc.IsMock {
		return dc.Mock.Get(ctx, key, dst)
	}
	return dc.Client.Get(ctx, key, dst)
}

func (dc *DatastoreClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	if dc.IsMock {
		return dc.Mock.GetAll(ctx, q, dst)
	}
	return dc.Client.GetAll(ctx, q, dst)
}

func (dc *DatastoreClient) NewQuery(kind string) *datastore.Query {
	return datastore.NewQuery(kind)
}

func (dc *DatastoreClient) NameKey(kind, name string, parent *datastore.Key) *datastore.Key {
	return datastore.NameKey(kind, name, parent)
}

func (dc *DatastoreClient) Delete(ctx context.Context, key *datastore.Key) error {
	if dc.IsMock {
		return dc.Mock.Delete(ctx, key)
	}
	return dc.Client.Delete(ctx, key)
}

func (dc *DatastoreClient) DeleteMulti(ctx context.Context, keys []*datastore.Key) error {
	if dc.IsMock {
		return dc.Mock.DeleteMulti(ctx, keys)
	}
	return dc.Client.DeleteMulti(ctx, keys)
}

func (dc *DatastoreClient) Count(ctx context.Context, q *datastore.Query) (int, error) {
	if dc.IsMock {
		return dc.Mock.Count(ctx, q)
	}
	return dc.Client.Count(ctx, q)
}

// Clear removes all data from the datastore (for testing)
func (dc *DatastoreClient) Clear() {
	if dc.IsMock && dc.Mock != nil {
		dc.Mock.Clear()
	}
}

// PutEntity stores any entity data (for testing)
func (dc *DatastoreClient) PutEntity(id string, entity interface{}) {
	if dc.IsMock && dc.Mock != nil {
		key := dc.NameKey("Entity", id, nil)
		dc.Mock.Put(context.Background(), key, entity)
	}
}
