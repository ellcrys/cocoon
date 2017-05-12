package archiver

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Object defines a struct for collecting objects to archive
type Object interface {

	// Read calls the callback with byte chunks of the object.
	// The implementation must stop and return any error returned by the callback
	Read(func(d []byte) error) error
}

// Persister provides a structure to archive an object
type Persister interface {

	// Write collects the data to persist
	Write(context.Context, []byte) error

	// Commit sends the data to the storage
	Commit() error
}

// Archiver provides a structure that collects and object and
// saves using a Persister implementation
type Archiver struct {
	object    Object
	persister Persister
}

// NewArchiver creates a new archiver
func NewArchiver(o Object, p Persister) *Archiver {
	return &Archiver{object: o, persister: p}
}

// Do archives the object using the persister implementation
func (a *Archiver) Do() error {

	if a.object == nil {
		return fmt.Errorf("object has not been set")
	}
	if a.persister == nil {
		return fmt.Errorf("persister has not been set")
	}

	// collect data from the object
	if err := a.object.Read(func(b []byte) error {
		if err := a.persister.Write(context.Background(), b); err != nil {
			return errors.Wrap(err, "failed to write to persister")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to read object")
	}

	// persist the data
	if err := a.persister.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit object")
	}

	return nil
}

// MakeArchiveName creates name to be used for archiving a cocooon code release
func MakeArchiveName(cocoonID, versionID string) string {
	return fmt.Sprintf("%s_%s.tar.gz", cocoonID, versionID)
}
