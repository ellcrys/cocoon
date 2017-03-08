package grpc

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ncodes/cocoon/core/api/grpc/proto"
	"github.com/ncodes/cocoon/core/scheduler"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("api.grpc")

// scheduler represents the cluster scheduler implementation (nomad, kubernetes, etc)
var sch scheduler.Scheduler

// SetScheduler sets the default cluster
func SetScheduler(s scheduler.Scheduler) {
	sch = s
}

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

// Deploy starts a new cocoon. The scheduler creates a job based on the requests
func (api *API) Deploy(ctx context.Context, req *proto.DeployRequest) (*proto.Response, error) {
	depInfo, err := sch.Deploy(
		req.GetId(),
		req.GetLanguage(),
		req.GetUrl(),
		req.GetReleaseTag(),
		string(req.GetBuildParam()),
		req.GetMemory(),
		req.GetCpuShare(),
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "system") {
			log.Error(err.Error())
			return nil, fmt.Errorf("failed to deploy cocoon")
		}
		return nil, err
	}

	log.Infof("Successfully deployed cocoon code %s", depInfo.ID)

	return &proto.Response{
		Status: 200,
		Body:   []byte(depInfo.ID),
	}, nil
}
