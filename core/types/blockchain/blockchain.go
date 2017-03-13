package blockchain

import (
	"github.com/ncodes/cocoon/core/types/store"
)

// Blockchain defines an interface for a blockchain
type Blockchain interface {
	Connect(dbAddr string) (interface{}, error)
	Init(name string) error
	CreateChain(name string, public bool) (*Chain, error)
	GetChain(name string) (*Chain, error)
	GetImplmentationName() string
	MakeChainName(namespace, name string) string
	CreateBlock(chainName string, transactions []*store.Transaction) (*Block, error)
	// GetBlock()
}
