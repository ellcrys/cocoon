package types

import "fmt"
import "strings"

// ErrIdentityNotFound indicates a non-existing identity
var ErrIdentityNotFound = fmt.Errorf("identity not found")

// ErrIdentityAlreadyExists indicates existence of an identity
var ErrIdentityAlreadyExists = fmt.Errorf("An identity with matching email already exists")

// ErrLedgerNotFound indicates a missing ledger
var ErrLedgerNotFound = fmt.Errorf("ledger not found")

// ErrTxNotFound indicates a missing transaction
var ErrTxNotFound = fmt.Errorf("transaction not found")

// ErrChainNotFound indicates a missing chain
var ErrChainNotFound = fmt.Errorf("chain not found")

// ErrZeroTransactions indicates a transaction list has zeto transactions
var ErrZeroTransactions = fmt.Errorf("zero transactions not allowed")

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
