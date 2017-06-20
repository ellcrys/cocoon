package api

import (
	"fmt"
	"net"
	"strings"
	"time"

	"os"

	"github.com/chuckpreslar/emission"
	"github.com/ellcrys/util"
	"github.com/ellcrys/cocoon/core/api/api/proto_api"
	"github.com/ellcrys/cocoon/core/config"
	"github.com/ellcrys/cocoon/core/platform"
	"github.com/ellcrys/cocoon/core/types"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var apiLog = config.MakeLogger("api.rpc")

// SetLogLevel sets the log level of the logger
func SetLogLevel(l logging.Level) {
	logging.SetLevel(l, apiLog.Module)
}

// API defines a GRPC api for performing various
// cocoon operations such as cocoon orchestration, resource
// allocation etc
type API struct {
	server       *grpc.Server
	endedCh      chan bool
	platform     *platform.Platform
	logProvider  types.LogProvider
	EventEmitter *emission.Emitter
}

// NewAPI creates a new GRPCAPI object
func NewAPI() (*API, error) {
	platform, err := platform.NewPlatform()
	if err != nil {
		return nil, err
	}
	eventEmitter := emission.NewEmitter()
	eventEmitter.SetMaxListeners(20)
	return &API{
		logProvider:  &StackDriverLog{},
		platform:     platform,
		EventEmitter: eventEmitter,
	}, nil
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

	time.AfterFunc(1*time.Second, func() {
		apiLog.Info("Server has started")
		apiLog.Infof("         RPC Port = %s", strings.Split(addr, ":")[1])
		apiLog.Infof("      Environment = %s", util.Env("ENV", "development"))
		apiLog.Infof("      API Version = %s", util.Env("API_VERSION", ""))
		apiLog.Infof("Connector Version = %s", util.Env("CONNECTOR_VERSION", ""))
		api.EventEmitter.Emit("started")
	})

	api.server = grpc.NewServer(grpc.UnaryInterceptor(api.Interceptors()))
	proto_api.RegisterAPIServer(api.server, api)
	api.server.Serve(lis)
}

// Stop stops the api and returns an exit code.
func (api *API) Stop() {
	api.server.Stop()
	api.platform.Stop()
	close(api.endedCh)
}
