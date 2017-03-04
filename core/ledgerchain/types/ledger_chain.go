package types

// LedgerChain defines an interface required for the implementation
// of a linked ledger data structure. A ledger chain is a continous,
// crytographically linked set of ledger with each ledger holding unlimited
// number of cryptographically linked transactions.
type LedgerChain interface {
	Connect(dbAddr string) (interface{}, error)
	Init(globalLedgerName string) error
	GetBackend() string
	CreateLedger(name string, public bool) (*Ledger, error)
	GetLedger(name string) (*Ledger, error)
	Put(txID, ledger, key, value string) (*Transaction, error)
	Get(ledger, key string) (*Transaction, error)
	GetByID(txID string) (*Transaction, error)
	MakeLedgerName(namespace, name string) string
	MakeTxKey(namespace, name string) string
	Close() error
}
