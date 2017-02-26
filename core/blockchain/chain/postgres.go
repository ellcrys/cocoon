package chain

import "errors"

// ErrChainExist represents an error about an existing chain
var ErrChainExist = errors.New("chain already exists")

// PostgresChain defines a blockchain implementation
// modelled on the postgres database. It implements
// the Chain interface
type PostgresChain struct {
}

// GetBackend returns the database backend this chain
// implementation depends on.
func (ch *PostgresChain) GetBackend() string {
	return "postgres"
}

// Connect connects to a postgress server and returns a client
// or error if connection failed.
func (ch *PostgresChain) Connect(dbAddr string) (interface{}, error) {
	return nil, nil
}

// CreateChain creates a new chain. This is represented as a table.
// Returns ErrChainExist if chain already exists or error for other
// types of issues or nil if all goes well.
func (ch *PostgresChain) CreateChain(name string) error {
	return nil
}
