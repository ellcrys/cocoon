package scheduler

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

// NomadServiceDiscovery provides service discovery to the nomad schedulerAddr
// by querying a consul server
type NomadServiceDiscovery struct {
	consulClient *api.Client
}

// NewNomadServiceDiscovery creates a nomad service discovery instance
func NewNomadServiceDiscovery(client *api.Client) *NomadServiceDiscovery {
	return &NomadServiceDiscovery{
		consulClient: client,
	}
}

// GetByID fetches all the addresses of a service by name.
func (d *NomadServiceDiscovery) GetByID(name string, query map[string]string) ([]*Service, error) {

	catalog := d.consulClient.Catalog()
	var tag = query["tag"]
	var dc = query["dc"]

	consulServices, _, err := catalog.Service(name, tag, &api.QueryOptions{Datacenter: dc})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %s", err)
	}

	var services []*Service
	for _, srv := range consulServices {
		serviceAddr := srv.ServiceAddress
		if len(serviceAddr) == 0 {
			serviceAddr = srv.Address
		}
		services = append(services, &Service{
			Name: srv.ServiceName,
			ID:   srv.ServiceID,
			IP:   serviceAddr,
			Port: srv.ServicePort,
			Tags: srv.ServiceTags,
		})
	}

	return services, nil
}
