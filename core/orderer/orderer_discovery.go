package orderer

import (
	"fmt"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/scheduler"
	logging "github.com/op/go-logging"
)

var discoveryLog = logging.MustGetLogger("orderer.discovery")

// Discovery defines a structure for fetching a list of addresses of orderers
// accessible in the cluster.
type Discovery struct {
	mu           *sync.Mutex
	orderersAddr []string
	ticker       *time.Ticker
	OnUpdateFunc func(addrs []string)
}

// NewDiscovery creates a new discovery object.
func NewDiscovery() *Discovery {
	return &Discovery{
		mu: &sync.Mutex{},
	}
}

// Discover fetches a list of orderer service addresses
// via consul service discovery API.
// For development purpose, If DEV_ORDERER_ADDR is set,
// it will fetch the orderer address from the env variable.
func (od *Discovery) discover() error {

	var err error

	if len(os.Getenv("DEV_ORDERER_ADDR")) > 0 {
		od.orderersAddr = []string{os.Getenv("DEV_ORDERER_ADDR")}
		return nil
	}

	ds := scheduler.NomadServiceDiscovery{
		ConsulAddr: util.Env("CONSUL_ADDR", "localhost:8500"),
		Protocol:   "http",
	}

	_orderers, err := ds.GetByID("orderers", nil)
	if err != nil {
		return err
	}

	var orderers []string
	for _, orderer := range _orderers {
		orderers = append(orderers, fmt.Sprintf("%s:%d", orderer.IP, int(orderer.Port)))
	}

	od.mu.Lock()
	od.orderersAddr = orderers
	od.mu.Unlock()
	return nil
}

// Discover starts a ticker that discovers and updates the list
// of orderer addresses. It will perform the discovery immediately
// and will return error if it fails, otherwise nil is returned and
// subsequent discovery will be performed periodically
func (od *Discovery) Discover() error {

	// run immediately
	if err := od.discover(); err != nil {
		return err
	}

	// run on interval
	od.ticker = time.NewTicker(15 * time.Second)
	for _ = range od.ticker.C {
		err := od.discover()
		if err != nil {
			discoveryLog.Error(err.Error())
			if od.OnUpdateFunc != nil {
				od.OnUpdateFunc(od.GetAddrs())
			}
		}
	}
	return nil
}

// GetAddrs returns the list of discovered addresses
func (od *Discovery) GetAddrs() []string {
	return od.orderersAddr
}

// GetGRPConn dials a random orderer address and returns a
// grpc connection. If no orderer address has been discovered, nil and are error are returned.
func (od *Discovery) GetGRPConn() (*grpc.ClientConn, error) {

	var selected string
	log.Info("Here")

	if len(od.orderersAddr) == 0 {
		od.mu.Lock()
		return nil, fmt.Errorf("no known orderer address")
	}
	log.Info("Here 3")
	if len(od.orderersAddr) == 1 {
		selected = od.orderersAddr[0]
	} else {
		selected = od.orderersAddr[util.RandNum(0, len(od.orderersAddr))]
	}
	log.Info("Here 4")
	client, err := grpc.Dial(selected, grpc.WithInsecure())
	if err != nil {
		od.mu.Unlock()
		return nil, err
	}

	od.mu.Unlock()
	log.Info("Here 5")

	return client, nil
}

// Stop stops the discovery ticker
func (od *Discovery) Stop() {
	od.ticker.Stop()
}
