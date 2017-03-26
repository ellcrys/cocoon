package types

import "fmt"
import "strings"

var (
	// ErrCocoonCodeNotRunning indicates a launched cocoon that isn't running a cocoon code
	ErrCocoonCodeNotRunning = fmt.Errorf("cocoon code is not running")

	// ErrIdentityNotFound indicates a non-existing identity
	ErrIdentityNotFound = fmt.Errorf("identity not found")

	// ErrIdentityAlreadyExists indicates existence of an identity
	ErrIdentityAlreadyExists = fmt.Errorf("an identity with matching email already exists")

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

	// ErrCocoonNotFound represents a missing/unknown cocoon
	ErrCocoonNotFound = fmt.Errorf("cocoon not found")

	// ErrInvalidOrExpiredToken represents an invalid/expired access token
	ErrInvalidOrExpiredToken = fmt.Errorf("access token is invalid or expired")

	// ErrClientNoActiveSession represents a lack of active user session on the client
	ErrClientNoActiveSession = fmt.Errorf("No active session. Please login")

	// ErrUninitializedStream represents a stream with nil value
	ErrUninitializedStream = fmt.Errorf("stream appears to be uninitialized")
)

// IsDuplicatePrevBlockHashError checks whether an error is one created when an ttempts to
// create a block with a prev block hash value already used.
func IsDuplicatePrevBlockHashError(err error) bool {
	return strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "idx_name_prev_block_hash"`)
}
