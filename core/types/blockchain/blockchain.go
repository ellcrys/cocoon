package blockchain

// Blockchain defines an interface for a blockchain
type Blockchain interface {
	Connect(dbAddr string) (interface{}, error)
	Init(name string) error
	CreateChain(name string, public bool) (*Chain, error)
	GetImplmentationName() string
	MakeChainName(namespace, name string) string
}
