package orderer

import (
	"fmt"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/ellcrys/cocoon/core/config"
	"github.com/ellcrys/cocoon/core/scheduler"
	"github.com/ellcrys/util"
	"github.com/hashicorp/consul/api"
)

var discoveryLog = config.MakeLogger("orderer.discovery")

// DiscoveryInterval is the time between each discovery checks
var DiscoveryInterval = time.Second * 5

// Discovery defines a structure for fetching a list of addresses of orderers
// accessible in the cluster.
type Discovery struct {
	sync.Mutex
	sd           scheduler.ServiceDiscovery
	consulClient *api.Client
	orderersAddr []string
	ticker       *time.Ticker
	OnUpdateFunc func(addrs []string)
}

// NewDiscovery creates a new discovery object.
// Returns error if unable to connector to the service discovery.
// Setting the env variable `CONSUL_ADDR` will override the default config address.
func NewDiscovery() (*Discovery, error) {
	cfg := api.DefaultConfig()
	cfg.Address = util.Env("CONSUL_ADDR", cfg.Address)
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}
	return &Discovery{
		sd:           scheduler.NewNomadServiceDiscovery(client),
		consulClient: client,
	}, nil
}

// Discover fetches a list of orderer service addresses
// via consul service discovery API.
// For development purpose, If DEV_ORDERER_ADDR is set,
// it will fetch the orderer address from the env variable.
func (od *Discovery) discover() error {

	var err error

	if len(os.Getenv("DEV_ORDERER_ADDR")) > 0 {
		od.Lock()
		defer od.Unlock()
		od.orderersAddr = []string{os.Getenv("DEV_ORDERER_ADDR")}
		return nil
	}

	_orderers, err := od.sd.GetByID("orderers", nil)
	if err != nil {
		return err
	}

	for _, orderer := range _orderers {
		od.Add(fmt.Sprintf("%s:%d", orderer.IP, int(orderer.Port)))
	}

	return nil
}

// Add a new orderer address
func (od *Discovery) Add(addr string) {
	od.Lock()
	defer od.Unlock()
	od.orderersAddr = append(od.orderersAddr, addr)
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
	od.ticker = time.NewTicker(DiscoveryInterval)
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

// addOrdererService adds an orderer service. This is meant to be used
// testing and never accessible beyond this package.
func (od *Discovery) addOrdererService(addr string, port int) error {
	_, err := od.consulClient.Catalog().Register(&api.CatalogRegistration{
		Address: "192.0.12.1",
		Node:    "abc",
		Service: &api.AgentService{
			Service: "orderers",
			Address: addr,
			Port:    port,
		},
	}, nil)
	return err
}

// GetAddrs returns the list of discovered addresses
func (od *Discovery) GetAddrs() []string {
	od.Lock()
	defer od.Unlock()
	return od.orderersAddr
}

// Len returns the number of orderer addresses
func (od *Discovery) Len() int {
	od.Lock()
	defer od.Unlock()
	return len(od.orderersAddr)
}

// GetRandAddr returns a randomly selected address or an empty
// string if no address is available
func (od *Discovery) GetRandAddr() string {
	od.Lock()
	defer od.Unlock()
	if nOrderer := len(od.orderersAddr); nOrderer > 0 {
		return od.orderersAddr[util.RandNum(0, nOrderer)]
	}
	return ""
}

// GetGRPConn dials a random orderer address and returns a
// grpc connection. If no orderer address has been discovered, nil and are error are returned.
func (od *Discovery) GetGRPConn() (*grpc.ClientConn, error) {

	var addr = od.GetRandAddr()

	if addr == "" {
		return nil, fmt.Errorf("no known orderer address")
	}

	client, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Stop stops the discovery ticker
func (od *Discovery) Stop() {
	od.ticker.Stop()
}
