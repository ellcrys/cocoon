package types

// Store defines an interface required for the implementation
// of a data structure for creating and reading ledgers and transactions.
type Store interface {
	Connect(dbAddr string) (interface{}, error)
	Init(systemPublicLedgerName, systemPrivateLedgerName string) error
	SetBlockchainImplementation(b Blockchain)
	GetImplementationName() string
	CreateLedger(name string, chained, public bool) (*Ledger, error)
	CreateLedgerThen(name string, chained, public bool, then func() error) (*Ledger, error)
	GetLedger(name string) (*Ledger, error)
	Put(ledger string, txs []*Transaction) ([]*TxReceipt, error)
	PutThen(ledger string, txs []*Transaction, then func(storedTxs []*Transaction) error) ([]*TxReceipt, error)
	Get(ledger, key string) (*Transaction, error)
	GetByID(ledger, txID string) (*Transaction, error)
	GetRange(ledger, startKey, endKey string, inclusive bool, limit, lastNum int) ([]*Transaction, error)
	Close() error
}
