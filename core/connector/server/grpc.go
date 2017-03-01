package server

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ncodes/cocoon/core/connector/server/proto"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("connector.api")

// APIServer defines a grpc server for
// invoking operations against cocoon code
type APIServer struct {
	server  *grpc.Server
	endedCh chan bool
}

// NewAPIServer creates a new grpc API server
func NewAPIServer() *APIServer {
	return new(APIServer)
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
