package types

// Blockchain defines an interface for a blockchain
type Blockchain interface {
	Connect(dbAddr string) (interface{}, error)
	Init(name string) error
	GetImplmentationName() string
	MakeChainName(namespace, name string) string
	CreateChain(name string, public bool) (*Chain, error)
	GetChain(name string) (*Chain, error)
	CreateBlock(id, chainName string, transactions []*Transaction) (*Block, error)
	GetBlock(chainName, id string) (*Block, error)
}
