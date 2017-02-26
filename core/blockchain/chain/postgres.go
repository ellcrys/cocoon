package chain

import (
	"errors"

	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// ErrChainExist represents an error about an existing chain
var ErrChainExist = errors.New("chain already exists")

// PostgresChain defines a blockchain implementation
// modelled on the postgres database. It implements
// the Chain interface
type PostgresChain struct {
	db *gorm.DB
}

// GetBackend returns the database backend this chain
// implementation depends on.
func (ch *PostgresChain) GetBackend() string {
	return "postgres"
}

// Connect connects to a postgress server and returns a client
// or error if connection failed.
func (ch *PostgresChain) Connect(dbAddr string) (interface{}, error) {
	var err error
	ch.db, err = gorm.Open("postgres", "host=localhost user=ned dbname=cocoonchain sslmode=disable password=")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain backend")
	}
	return ch.db, nil
}

// CreateChain creates a new chain. This is represented as a table.
// Returns ErrChainExist if chain already exists or error for other
// types of issues or nil if all goes well.
func (ch *PostgresChain) CreateChain(name string) error {
	return nil
}

// Close releases any resource held
func (ch *PostgresChain) Close() error {
	if ch.db != nil {
		return ch.Close()
	}
	return nil
}
