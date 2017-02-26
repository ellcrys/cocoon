package orderer

import (
	"fmt"
	"net"

	context "golang.org/x/net/context"

	"time"

	"github.com/ncodes/cocoon/core/blockchain/chain"
	"github.com/ncodes/cocoon/core/blockchain/orderer/proto"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("orderer")

// Orderer defines a transaction ordering, block creation
// and inclusion module
type Orderer struct {
	server *grpc.Server
	chain  chain.Chain
}

// NewOrderer creates a new Orderer object
func NewOrderer() *Orderer {
	return new(Orderer)
}

// Start starts the order service
func (od *Orderer) Start(port string, startedCh, endedCh chan bool) {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", port, err)
	}

	time.AfterFunc(2*time.Second, func() {
		log.Info("GRPC server started with no wahala")
		startedCh <- true
	})

	od.server = grpc.NewServer()
	proto.RegisterOrdererServer(od.server, od)
	od.server.Serve(lis)
}

// SetChain sets the blockchain implementation to use.
func (od *Orderer) SetChain(ch chain.Chain) {
	log.Infof("Setting blockchain backend to %s", ch.GetBackend())
	od.chain = ch
}

// Put adds a new record to the chain
func (od *Orderer) Put(ctx context.Context, tx *proto.OrdererTx) (*proto.Response, error) {
	return nil, nil
}
