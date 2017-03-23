package scheduler

import (
	"fmt"

	"net/url"

	"github.com/ellcrys/util"
	"github.com/franela/goreq"
)

// NomadServiceDiscovery provides service discovery to the nomad schedulerAddr
// by querying a consul server
type NomadServiceDiscovery struct {
	ConsulAddr string
	Protocol   string
}

func (nsd *NomadServiceDiscovery) getAddr() string {
	addr := "http://" + nsd.ConsulAddr
	if nsd.Protocol == "https" {
		addr = "https://" + nsd.ConsulAddr
	}
	return addr
}

// GetByID fetches a services instances by the service id.
func (nsd *NomadServiceDiscovery) GetByID(name string, query map[string]string) ([]*Service, error) {

	item := url.Values{}
	for key, val := range query {
		item.Set(key, val)
	}

	addr := nsd.getAddr()
	req, err := goreq.Request{
		Uri:         fmt.Sprintf("%s/%s", addr, "v1/catalog/service/"+name),
		QueryString: item,
	}.Do()

	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	body, _ := req.Body.ToString()
	if req.StatusCode != 200 {
		return nil, fmt.Errorf(body)
	}

	var _services []map[string]interface{}
	err = util.FromJSON([]byte(body), &_services)
	if err != nil {
		return nil, err
	}

	var services []*Service
	for _, srv := range _services {
		serviceAddr := ""
		if srv["ServiceAddress"] == nil {
			serviceAddr = srv["Address"].(string)
		} else {
			serviceAddr = srv["ServiceAddress"].(string)
		}
		services = append(services, &Service{
			Name: srv["ServiceName"].(string),
			ID:   srv["ID"].(string),
			IP:   serviceAddr,
			Port: fmt.Sprintf("%d", srv["ServicePort"]),
			Tags: srv["ServiceTags"].([]string),
		})
	}

	return services, nil
}
