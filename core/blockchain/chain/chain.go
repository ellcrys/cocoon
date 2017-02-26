package chain

// Chain defines an interface for creating and accessing a blockchain
type Chain interface {
	Connect(dbAddr string) (interface{}, error)
	GetBackend() string
	CreateChain(name string) error
	Close() error
}
