package types

const (
	// TxCreateLedger represents a message to create a ledger
	TxCreateLedger = "CREATE_LEDGER"

	// TxPut represents a message to create a transaction
	TxPut = "PUT"

	// TxGetLedger represents a message to get a ledger
	TxGetLedger = "GET_LEDGER"

	// TxGet represents a message to get a transaction
	TxGet = "GET"

	// TxGetByID represents a message to get a transaction by id
	TxGetByID = "GET_BY_ID"

	// TxGetBlockByID represents a message to get ledger's block by id
	TxGetBlockByID = "GET_BLOCK_BY_ID"

	// TxRangeGet represents a message to get a range of transactions
	TxRangeGet = "RANGE_GET"
)
