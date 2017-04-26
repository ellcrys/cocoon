package router

import (
	"fmt"

	"os"

	"github.com/ellcrys/util"
	"github.com/hashicorp/consul/api"
	logging "github.com/op/go-logging"
)

// RouterDomain is the router's domain
var RouterDomain = util.Env("ROUTER_DOMAIN", "whatbay.co")

// Helper defines a structure for hooking up to the
// reverse proxy tool within the cluster. Current implementation
// is designed to add frontend and backend entries in consul which is a
// traefik backend
type Helper struct {
	client *api.Client
	l      *logging.Logger
}

// NewHelper creates a new router helper object. Returns error
// if unable to connector to consul
func NewHelper(l *logging.Logger) (*Helper, error) {
	cfg := api.DefaultConfig()
	fmt.Println(">>>> ", cfg.Address)
	fmt.Println(">2: ", util.Env("CONSUL_ADDR", cfg.Address), len(os.Getenv("CONSUL_ADDR")))
	cfg.Address = util.Env("CONSUL_ADDR", cfg.Address)
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}
	return &Helper{
		client: client,
		l:      l,
	}, nil
}

// AddFrontend adds a frontend to receive traffic from public internet
func (h *Helper) AddFrontend(name string) error {
	var frontend = fmt.Sprintf("/traefik/frontends/%s", name)
	var backendName = fmt.Sprintf("%s", name)
	var keys = map[string]string{
		frontend + "/backend":          backendName,
		frontend + "/entrypoints/0":    "http",
		frontend + "/routes/main/rule": fmt.Sprintf("Host:%s.%s", name, RouterDomain),
	}

	fmt.Println("Keys To Add: ")
	util.Printify(keys)

	kv := h.client.KV()
	var ops api.KVTxnOps
	for key, value := range keys {
		ops = append(ops, &api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   key,
			Value: []byte(value),
		})
	}

	ok, txResp, _, err := kv.Txn(ops, nil)
	if err != nil {
		return fmt.Errorf("failed to add frontend: %s", err)
	}
	fmt.Println("OK? ", ok)
	util.Printify(txResp)
	if ok {
		return nil
	}
	return fmt.Errorf("failed to add frontend")
}

// AddBackend adds a backend server. This method
// will overwrite any backend rule matching the name
// and include a server that is issued a random name.
func (h *Helper) AddBackend(name string) error {
	var backend = fmt.Sprintf("/traefik/backends/%s", name)
	var backendServer = fmt.Sprintf("%s/servers/%s_server", backend, name)
	var keys = map[string]string{
		backend + "/loadbalancer/method": "drr",
		backendServer + "/url":           "http://localhost:3000",
		backendServer + "/weight":        "10",
	}

	kv := h.client.KV()
	var ops api.KVTxnOps
	for key, value := range keys {
		ops = append(ops, &api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   key,
			Value: []byte(value),
		})
	}

	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return fmt.Errorf("failed to add backend: %s", err)
	}

	if ok {
		return nil
	}
	return fmt.Errorf("failed to add backend")
}
