package db

import (
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
