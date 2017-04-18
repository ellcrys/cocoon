package api

import (
	"fmt"
	"net"
	"strings"
	"time"

	"os"

	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/types"
	"google.golang.org/grpc"
)

var apiLog = config.MakeLogger("api.rpc", "api")

// API defines a GRPC api for performing various
// cocoon operations such as cocoon orchestration, resource
// allocation etc
type API struct {
	server           *grpc.Server
	endedCh          chan bool
	ordererDiscovery *orderer.Discovery
	scheduler        scheduler.Scheduler
	logProvider      types.LogProvider
}

// NewAPI creates a new GRPCAPI object
func NewAPI(scheduler scheduler.Scheduler) *API {
	return &API{
		scheduler:        scheduler,
		ordererDiscovery: orderer.NewDiscovery(),
		logProvider:      &StackDriverLog{},
	}
}

// Start starts the server
func (api *API) Start(addr string, endedCh chan bool) {

	api.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		apiLog.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	err = api.logProvider.Init(map[string]interface{}{"projectId": os.Getenv("GCP_PROJECT_ID")})
	if err != nil {
		apiLog.Fatalf("failed to initialize log provider: %v", err)
	}

	time.AfterFunc(2*time.Second, func() {
		apiLog.Infof("Started server on port %s", strings.Split(addr, ":")[1])
		go api.ordererDiscovery.Discover()
		time.Sleep(1 * time.Second)
		if len(api.ordererDiscovery.GetAddrs()) == 0 {
			apiLog.Warning("No orderer address was found. We won't be able to reach the orderer. ")
		}
	})

	api.server = grpc.NewServer()
	proto_api.RegisterAPIServer(api.server, api)
	api.server.Serve(lis)
}

// Stop stops the api and returns an exit code.
func (api *API) Stop(exitCode int) int {
	api.server.Stop()
	close(api.endedCh)
	return exitCode
}
