package client

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/config"
)

// APIAddress is the remote address to the cluster server
var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8004")

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}

// ValidateFirewall parses and validates the content of
// a firewall ruleset. It expects a json string or a slice
// of map[string]strins. It will return a slice of map[string]string
// values that represents valid firewall rules. Destination host addresses
// are not resolved.
func ValidateFirewall(firewall interface{}) ([]map[string]string, []error) {

	var errs []error
	if firewall == nil {
		errs = append(errs, fmt.Errorf("function value is nil"))
		return nil, errs
	}

	var firewallMap []map[string]string
	switch fwData := firewall.(type) {
	case string:
		if len(fwData) == 0 {
			errs = append(errs, fmt.Errorf("empty string passed"))
			return nil, errs
		}
		err := util.FromJSON([]byte(fwData), &firewallMap)
		if err != nil {
			errs = append(errs, fmt.Errorf("malformed json"))
			return nil, errs
		}
	case []map[string]string:
		firewallMap = fwData
	default:
		errs = append(errs, fmt.Errorf("invalid type. expected a json string or a slice of map"))
		return nil, errs
	}

	for i, rule := range firewallMap {
		if rule["destination"] == "" {
			errs = append(errs, fmt.Errorf("rule %d: destination is required", i))
		}
		if rule["port"] == "" {
			errs = append(errs, fmt.Errorf("rule %d: port is required", i))
		}
		if rule["protocol"] == "" {
			rule["protocol"] = "tcp"
		}
	}

	return firewallMap, errs
}
