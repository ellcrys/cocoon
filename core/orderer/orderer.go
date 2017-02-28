package orderer

import (
	"fmt"
	"net"
	"strings"

	context "golang.org/x/net/context"

	"time"

	"github.com/ncodes/cocoon/core/ledgerchain/types"
	"github.com/ncodes/cocoon/core/orderer/proto"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("orderer")

// Orderer defines a transaction ordering, block creation
// and inclusion module
type Orderer struct {
	server  *grpc.Server
	chain   types.LedgerChain
	endedCh chan bool
}

// NewOrderer creates a new Orderer object
func NewOrderer() *Orderer {
	return new(Orderer)
}

// Start starts the order service
func (od *Orderer) Start(addr, ledgerChainConStr string, endedCh chan bool) {

	od.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	time.AfterFunc(2*time.Second, func() {

		log.Infof("Started orderer GRPC server on port %s", strings.Split(addr, ":")[1])

		// establish connection to chain backend
		_, err := od.chain.Connect(ledgerChainConStr)
		if err != nil {
			log.Info(err)
			od.Stop(1)
			return
		}

		log.Info("Backend successfully connnected")
	})

	od.server = grpc.NewServer()
	proto.RegisterOrdererServer(od.server, od)
	od.server.Serve(lis)
}

// Stop stops the orderer and returns an exit code.
func (od *Orderer) Stop(exitCode int) int {
	od.server.Stop()
	od.chain.Close()
	close(od.endedCh)
	return exitCode
}

// SetLedgerChain sets the ledgerchain implementation to use.
func (od *Orderer) SetLedgerChain(ch types.LedgerChain) {
	log.Infof("Setting ledgerchain backend to %s", ch.GetBackend())
	od.chain = ch
}

// Put adds a new record to the chain
func (od *Orderer) Put(ctx context.Context, tx *proto.OrdererTx) (*proto.Response, error) {
	return nil, nil
}
