package types

import "fmt"
import "strings"

var (
	// ErrCocoonCodeNotRunning indicates a launched cocoon that isn't running a cocoon code
	ErrCocoonCodeNotRunning = fmt.Errorf("cocoon code is not running")

	// ErrIdentityNotFound indicates a non-existing identity
	ErrIdentityNotFound = fmt.Errorf("identity not found")

	// ErrIdentityAlreadyExists indicates existence of an identity
	ErrIdentityAlreadyExists = fmt.Errorf("An identity with matching email already exists")

	// ErrLedgerNotFound indicates a missing ledger
	ErrLedgerNotFound = fmt.Errorf("ledger not found")

	// ErrTxNotFound indicates a missing transaction
	ErrTxNotFound = fmt.Errorf("transaction not found")

	// ErrChainNotFound indicates a missing chain
	ErrChainNotFound = fmt.Errorf("chain not found")

	// ErrZeroTransactions indicates a transaction list has zeto transactions
	ErrZeroTransactions = fmt.Errorf("zero transactions not allowed")

	// ErrBlockNotFound indicates a missing block
	ErrBlockNotFound = fmt.Errorf("block not found")

	// ErrOperationTimeout represents a timeout error that occurs when response
	// is not received from orderer in time.
	ErrOperationTimeout = fmt.Errorf("operation timed out")
)

// IsPrevBlockConcurrentAccessError checks whether an error is one created when
// attempt is made to access a previous block that is currently being updated by a concurrent transaction.
func IsPrevBlockConcurrentAccessError(err error) bool {
	return err.Error() == "failed to update previous block right sibling column. pq: could not serialize access due to concurrent update"
}

// IsDuplicatePrevHashError checks whether an error is one created when a transaction attempts to
// create a record with a prev hash value already used by a previous commited transaction.
func IsDuplicatePrevHashError(err error) bool {
	return strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "idx_name_prev_hash"`)
}
