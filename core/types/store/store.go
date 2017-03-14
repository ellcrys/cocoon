package store

// Store defines an interface required for the implementation
// of a data structure for creating and reading ledgers and transactions.
type Store interface {
	Connect(dbAddr string) (interface{}, error)
	Init(globalLedgerName string) error
	GetImplmentationName() string
	CreateLedger(name string, chained, public bool) (*Ledger, error)
	CreateLedgerThen(name string, chained, public bool, then func() error) (*Ledger, error)
	GetLedger(name string) (*Ledger, error)
	Put(txID, ledger, key, value string) (*Transaction, error)
	Get(ledger, key string) (*Transaction, error)
	GetByID(txID string) (*Transaction, error)
	MakeLedgerName(namespace, name string) string
	MakeTxKey(namespace, name string) string
	Close() error
}
