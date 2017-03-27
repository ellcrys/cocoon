package server

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ncodes/cocoon/core/connector"
	"github.com/ncodes/cocoon/core/connector/server/proto"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("connector.rpc")

// RPCServer defines a grpc server for
// invoking operations against cocoon code
type RPCServer struct {
	server    *grpc.Server
	endedCh   chan bool
	connector *connector.Connector
}

// NewRPCServer creates a new grpc API server
func NewRPCServer(connector *connector.Connector) *RPCServer {
	server := new(RPCServer)
	server.connector = connector
	return server
}

// Start starts the API service
func (rpc *RPCServer) Start(addr string, startedCh chan bool, endedCh chan bool) {

	rpc.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	time.AfterFunc(2*time.Second, func() {
		log.Infof("Started GRPC API server on port %s", strings.Split(addr, ":")[1])
		startedCh <- true
	})

	rpc.server = grpc.NewServer()
	proto.RegisterRPCServer(rpc.server, rpc)
	rpc.server.Serve(lis)
}

// Stop stops the orderer and returns an exit capie.
func (rpc *RPCServer) Stop(exitCode int) int {
	rpc.server.Stop()
	close(rpc.endedCh)
	return exitCode
}

// Invoke calls a function in the cocoon code.
func (rpc *RPCServer) Invoke(ctx context.Context, req *proto.InvokeRequest) (*proto.InvokeResponse, error) {
	log.Infof("New invoke transaction (%s)", req.GetId())

	// var respCh = make(chan *stub_proto.Tx)
	// var txID = req.GetId()
	// err := rpc.connector.GetClient().SendTx(&stub_proto.Tx{
	// 	Id: txID,
	// 	// Invoke: true,
	// 	Name:   "function",
	// 	Params: append([]string{req.GetFunction()}, req.GetParams()...),
	// }, respCh)

	// if err != nil {
	// 	log.Debugf("Failed to send transaction [%s] to cocoon code. %s", txID, err)
	// 	return nil, err
	// }

	// resp, err := common.AwaitTxChan(respCh)
	// if err != nil {
	// 	return nil, err
	// }

	// return &proto.InvokeResponse{
	// 	Id:       txID,
	// 	Function: req.GetFunction(),
	// 	Body:     resp.GetBody(),
	// }, nil
	return nil, nil
}
