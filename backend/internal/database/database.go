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