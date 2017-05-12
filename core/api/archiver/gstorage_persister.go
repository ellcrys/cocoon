package archiver

import (
	"fmt"
	"os"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// PersisterProjectID is the google cloud project id
var PersisterProjectID = os.Getenv("GCP_PROJECT_ID")

// BucketName is the name of the google cloud repo
var BucketName = os.Getenv("REPO_ARCHIVE_BKT")

// GStoragePersister defines a structure that implements
// the Persister interface for storing objects on google cloud storage.
type GStoragePersister struct {
	name         string
	client       *storage.Client
	bucket       *storage.BucketHandle
	object       *storage.ObjectHandle
	objectWriter *storage.Writer
}

// NewGStoragePersister creates a new google storage persister
func NewGStoragePersister(name string) (*GStoragePersister, error) {

	var err error

	if PersisterProjectID == "" {
		return nil, fmt.Errorf("project id not set")
	}
	if BucketName == "" {
		return nil, fmt.Errorf("bucket name not set")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	return &GStoragePersister{
		name:   name,
		client: client,
		bucket: client.Bucket(BucketName),
		object: nil,
	}, nil
}

// Write data into the object
func (g *GStoragePersister) Write(ctx context.Context, p []byte) error {

	if g.object == nil {
		g.object = g.bucket.Object(g.name)
		g.objectWriter = g.object.NewWriter(ctx)
	}

	_, err := g.objectWriter.Write(p)
	if err != nil {
		return err
	}

	return nil
}

// Commit saves the object
func (g *GStoragePersister) Commit() error {
	if g.objectWriter == nil {
		return fmt.Errorf("writer not initialized")
	}
	return g.objectWriter.Close()
}
