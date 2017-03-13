package blockchain

// Blockchain defines an interface for a blockchain
type Blockchain interface {
	Connect(dbAddr string) (interface{}, error)
	Init(name string) error
	Create(name string) error
}
