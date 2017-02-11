package orderer

import (
	"github.com/ncodes/cocoon/core/blockchain"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("orderer")

// Orderer defines a transaction ordering, block creation
// and inclusion module
type Orderer struct {
}

// NewOrderer creates a new Orderer object
func NewOrderer() *Orderer {
	return new(Orderer)
}

// Start starts the order service
func (od *Orderer) Start(close chan bool, txnChan chan blockchain.Transaction) {
	log.Info("Orderer has started")
}
