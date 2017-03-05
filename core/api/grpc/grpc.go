package grpc

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ncodes/cocoon/core/api/grpc/proto"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("api.grpc")

// API defines a GRPC api for performing various
// cocoon operations such as cocoon orchestration, resource
// allocation etc
type API struct {
	server  *grpc.Server
	endedCh chan bool
}

// NewAPI creates a new GRPCAPI object
func NewAPI() *API {
	return new(API)
}

// Start starts the server
func (api *API) Start(addr string, endedCh chan bool) {

	api.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	time.AfterFunc(2*time.Second, func() {
		log.Infof("Started server on port %s", strings.Split(addr, ":")[1])
	})

	api.server = grpc.NewServer()
	proto.RegisterAPIServer(api.server, api)
	api.server.Serve(lis)
}

// Stop stops the api and returns an exit code.
func (api *API) Stop(exitCode int) int {
	api.server.Stop()
	close(api.endedCh)
	return exitCode
}

// Cocoon handles cocoon requests.
func (api *API) Cocoon(context.Context, *proto.Request) (*proto.Response, error) {
	return nil, nil
}
