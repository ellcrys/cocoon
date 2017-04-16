package server

import (
	"fmt"
	"net"
	"time"

	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector"
	"github.com/ncodes/cocoon/core/connector/server/connector_proto"
	"github.com/ncodes/cocoon/core/connector/server/handlers"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log *logging.Logger

// RPCServer defines a grpc server for
// invoking operations against cocoon code
type RPCServer struct {
	server        *grpc.Server
	connector     *connector.Connector
	ledgerOps     *handlers.LedgerOperations
	cocoonCodeOps *handlers.CocoonCodeOperations
}

// NewRPCServer creates a new grpc API server
func NewRPCServer(connector *connector.Connector) *RPCServer {
	log = config.MakeLogger("connector.rpc", fmt.Sprintf("cocoon.%s", connector.GetRequest().ID))
	server := new(RPCServer)
	server.connector = connector
	server.ledgerOps = handlers.NewLedgerOperationHandler(log, connector)
	server.cocoonCodeOps = handlers.NewCocoonCodeHandler("127.0.0.1" + connector.GetCocoonCodeRPCAddr())
	return server
}

// GetConnector returns the connector
func (rpc *RPCServer) GetConnector() *connector.Connector {
	return rpc.connector
}

// Start starts the API service
func (rpc *RPCServer) Start(addr string, startedCh chan bool) {

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
	connector_proto.RegisterConnectorServer(rpc.server, rpc)
	rpc.server.Serve(lis)
	rpc.Stop(1)
}

// Stop stops the orderer and returns an exit code.
func (rpc *RPCServer) Stop(exitCode int) int {
	if rpc.server != nil {
		rpc.server.Stop()
	}
	return exitCode
}

// Transact handles cocoon code or ledger bound transactions.
func (rpc *RPCServer) Transact(ctx context.Context, req *connector_proto.Request) (*connector_proto.Response, error) {
	switch req.OpType {
	case connector_proto.OpType_LedgerOp:
		return rpc.ledgerOps.Handle(ctx, req.LedgerOp)
	case connector_proto.OpType_CocoonCodeOp:
		return rpc.cocoonCodeOps.Handle(ctx, req.CocoonCodeOp)
	default:
		return nil, fmt.Errorf("unsupported operation type")
	}
}
