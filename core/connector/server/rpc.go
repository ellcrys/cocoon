package server

import (
	"fmt"
	"net"
	"time"

	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/connector"
	"github.com/ncodes/cocoon/core/connector/server/handlers"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log *logging.Logger

// RPC defines a grpc server for
// invoking a cocoon code
type RPC struct {
	server        *grpc.Server
	connector     *connector.Connector
	ledgerOps     *handlers.LedgerOperations
	cocoonCodeOps *handlers.CocoonCodeOperations
	lockOps       *handlers.LockOperations
}

// NewRPC creates a new grpc API server
func NewRPC(connector *connector.Connector) *RPC {
	log = config.MakeLogger("connector.rpc", "")
	server := new(RPC)
	server.connector = connector
	server.ledgerOps = handlers.NewLedgerOperationHandler(log, connector)
	server.cocoonCodeOps = handlers.NewCocoonCodeHandler(connector.GetCocoonCodeRPCAddr())
	server.lockOps = handlers.NewLockOperationHandler(log, connector)
	return server
}

// GetConnector returns the connector
func (rpc *RPC) GetConnector() *connector.Connector {
	return rpc.connector
}

// Start starts the API service
func (rpc *RPC) Start(addr string, startedCh chan bool) {

	_, port, _ := net.SplitHostPort(addr)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", port, err)
	}

	time.AfterFunc(2*time.Second, func() {
		log.Infof("Started GRPC API server @ %s", addr)
		startedCh <- true
		time.Sleep(1 * time.Second)
	})

	rpc.server = grpc.NewServer()
	proto_connector.RegisterConnectorServer(rpc.server, rpc)
	rpc.server.Serve(lis)
}

// Stop stops the server
func (rpc *RPC) Stop() {
	rpc.stopCocoonCode()
	if rpc.server != nil {
		rpc.server.Stop()
	}
}

// Transact handles cocoon code or ledger bound transactions.
func (rpc *RPC) Transact(ctx context.Context, req *proto_connector.Request) (*proto_connector.Response, error) {
	switch req.OpType {
	case proto_connector.OpType_LedgerOp:
		return rpc.ledgerOps.Handle(ctx, req.LedgerOp)
	case proto_connector.OpType_LockOp:
		return rpc.lockOps.Handle(ctx, req.LockOp)
	default:
		return nil, fmt.Errorf("unsupported operation type")
	}
}

// stop the cocoon code
func (rpc *RPC) stopCocoonCode() {
	ctx, cc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cc()
	err := rpc.cocoonCodeOps.Stop(ctx)
	if err != nil {
		log.Errorf("failed to call Stop() on cocoon code: %s", err.Error())
		return
	}
	log.Info("Called Stop() on cocoon code")
}
