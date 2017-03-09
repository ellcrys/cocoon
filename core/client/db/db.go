package db

import (
	"bytes"
	"log"
	"path"

	"github.com/boltdb/bolt"
	homedir "github.com/mitchellh/go-homedir"
)

var defaultDB *bolt.DB

// ProjectName is the official name of the project
var ProjectName = "cocoon"

// GetDefaultDB returns a handle to the client's database
func GetDefaultDB() *bolt.DB {

	if defaultDB != nil {
		return defaultDB
	}

	home, _ := homedir.Dir()
	dbFile := path.Join(home, ".config", ProjectName, "client.db")
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	defaultDB = db
	return defaultDB
}

// GetFirstByPrefix returns the first key with the matching prefix
func GetFirstByPrefix(db *bolt.DB, bucket, prefix string) ([]byte, []byte, error) {
	var k, v []byte
	err := db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		for _k, _v := c.Seek([]byte(prefix)); _k != nil && bytes.HasPrefix(_k, []byte(prefix)); _k, _v = c.Next() {
			k = _k
			v = _v
			return nil
		}
		return nil
	})
	return k, v, err
}
