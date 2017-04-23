package types

const (
	// TxNewLedger represents a message to create a ledger
	TxNewLedger = "CREATE_LEDGER"

	// TxPut represents a message to create a transaction
	TxPut = "PUT"

	// TxGetLedger represents a message to get a ledger
	TxGetLedger = "GET_LEDGER"

	// TxGet represents a message to get a transaction
	TxGet = "GET"

	// TxGetBlockByID represents a message to get ledger's block by id
	TxGetBlockByID = "GET_BLOCK_BY_ID"

	// TxRangeGet represents a message to get a range of transactions
	TxRangeGet = "RANGE_GET"

	// OpLockAcquire represents a message to acquire a lock
	OpLockAcquire = "LOCK_ACQUIRE"

	// OpLockRelease represents a message to release a lock
	OpLockRelease = "LOCK_RELEASE"

	// OpLockCheckAcquire represents a message to check whether a session is still the acquirer of a lock
	OpLockCheckAcquire = "LOCK_CHECK_ACQUIRE"
)
