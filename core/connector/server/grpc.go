package server

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/connector/client"
	"github.com/ncodes/cocoon/core/connector/server/proto"
	stub "github.com/ncodes/cocoon/core/stubs/golang"
	stub_proto "github.com/ncodes/cocoon/core/stubs/golang/proto"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("connector.api")

// APIServer defines a grpc server for
// invoking operations against cocoon code
type APIServer struct {
	server           *grpc.Server
	endedCh          chan bool
	cocoonCodeClient *client.Client
}

// NewAPIServer creates a new grpc API server
func NewAPIServer(cocoonCodeClient *client.Client) *APIServer {
	server := new(APIServer)
	server.cocoonCodeClient = cocoonCodeClient
	return server
}

// Start starts the API service
func (api *APIServer) Start(addr string, endedCh chan bool) {

	api.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	time.AfterFunc(2*time.Second, func() {
		log.Infof("Started GRPC API server on port %s", strings.Split(addr, ":")[1])
	})

	api.server = grpc.NewServer()
	proto.RegisterAPIServer(api.server, api)
	api.server.Serve(lis)
}

// Stop stops the orderer and returns an exit capie.
func (api *APIServer) Stop(exitCode int) int {
	api.server.Stop()
	close(api.endedCh)
	return exitCode
}

// Invoke calls a function in the cocoon code.
func (api *APIServer) Invoke(ctx context.Context, req *proto.InvokeRequest) (*proto.InvokeResponse, error) {
	log.Infof("New invoke transaction (%s)", req.GetId())

	var respCh = make(chan *stub_proto.Tx)
	var txID = util.UUID4()
	err := api.cocoonCodeClient.SendTx(&stub_proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   "function",
		Params: append([]string{req.GetFunction()}, req.GetParams()...),
	}, respCh)

	if err != nil {
		log.Debugf("Failed to send transaction [%s] to cocoon code. %s", txID, err)
		return nil, err
	}

	resp, err := stub.AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}

	log.Infof("Resp: %v", resp)

	return nil, fmt.Errorf("test error")
}
