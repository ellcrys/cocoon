package api

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/orderer"
	"github.com/ncodes/cocoon/core/scheduler"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("api.grpc")

// API defines a GRPC api for performing various
// cocoon operations such as cocoon orchestration, resource
// allocation etc
type API struct {
	server           *grpc.Server
	endedCh          chan bool
	orderDiscoTicker *time.Ticker
	ordererAddrs     []string
	scheduler        scheduler.Scheduler
}

// NewAPI creates a new GRPCAPI object
func NewAPI(scheduler scheduler.Scheduler) *API {
	return &API{
		scheduler: scheduler,
	}
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

		var ordererAddrs []string
		ordererAddrs, err := orderer.DiscoverOrderers()
		if err != nil {
			log.Fatalf("failed to discover orderer. %s", err)
		}
		api.ordererAddrs = ordererAddrs

		if len(api.ordererAddrs) > 0 {
			log.Infof("Orderer address list updated. Contains %d orderer address(es)", len(api.ordererAddrs))
			return
		}

		log.Warning("No orderer address was found. We won't be able to reach the orderer. ")
	})

	// start a ticker to continously discover orderer addresses
	go func() {
		api.orderDiscoTicker = time.NewTicker(60 * time.Second)
		for _ = range api.orderDiscoTicker.C {
			ordererAddrs, err := orderer.DiscoverOrderers()
			if err != nil {
				log.Error(err)
				continue
			}
			api.ordererAddrs = ordererAddrs
		}
	}()

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
