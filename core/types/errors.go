package types

import "fmt"

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
