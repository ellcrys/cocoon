package api

import (
	"fmt"

	"strings"

	"github.com/asaskevich/govalidator"
	cocoon_util "github.com/ellcrys/cocoon-util"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/connector/server/acl"
	"github.com/ellcrys/cocoon/core/scheduler"
	"github.com/ellcrys/cocoon/core/types"
	"github.com/ellcrys/util"
)

// ValidateCocoon validates a cocoon to be created
func ValidateCocoon(c *types.Cocoon) error {

	if len(c.ID) == 0 {
		return fmt.Errorf("id: id is required")
	}
	if !common.IsValidCocoonID(c.ID) {
		return fmt.Errorf("id: id is not a valid resource name")
	}
	if c.Memory == 0 {
		return fmt.Errorf("resources.memory: memory is required")
	}
	if c.CPUShare == 0 {
		return fmt.Errorf("resources.cpuShare: CPU share is required")
	}
	if common.GetResourceSet(c.Memory, c.CPUShare) == nil {
		return fmt.Errorf("resources: Unknown resource set")
	}
	if c.NumSignatories <= 0 {
		return fmt.Errorf("signatories.max: number of signatories cannot be less than 1")
	}
	if c.SigThreshold <= 0 {
		return fmt.Errorf("signatories.threshold: signatory threshold cannot be less than 1")
	}
	if c.SigThreshold > c.NumSignatories {
		return fmt.Errorf("signatories.threshold: signatory threshold cannot be greater than maximum number of signatories")
	}
	if c.NumSignatories < len(c.Signatories) {
		return fmt.Errorf("signatories.signatories: max signatories already added. You can't add more")
	}

	return nil
}

// ValidateRelease checks whether a release field values
// are valid.
func ValidateRelease(r *types.Release) error {
	if len(r.ID) == 0 {
		return fmt.Errorf("id is required")
	}
	if !govalidator.IsUUIDv4(r.ID) {
		return fmt.Errorf("id is not a valid uuid")
	}
	if len(r.CocoonID) == 0 {
		return fmt.Errorf("cocoon id is required")
	}
	if len(r.URL) == 0 {
		return fmt.Errorf("url is required")
	}
	if !cocoon_util.IsGithubRepoURL(r.URL) {
		return fmt.Errorf("url is not a valid github repo url")
	}
	if len(r.Language) == 0 {
		return fmt.Errorf("language is required")
	}
	if !util.InStringSlice(scheduler.SupportedCocoonCodeLang, r.Language) {
		return fmt.Errorf("language is not supported. Expects one of these values %s", scheduler.SupportedCocoonCodeLang)
	}
	if len(r.BuildParam) > 0 {
		var _r map[string]interface{}
		if util.FromJSON([]byte(r.BuildParam), &_r) != nil {
			return fmt.Errorf("build parameter is not valid json")
		}
	}
	if r.Firewall != nil {
		_, errs := ValidateFirewallRules(r.Firewall.ToMap())
		if len(errs) != 0 {
			return fmt.Errorf("firewall: %s, ", errs[0])
		}
	}
	if len(r.ACL) > 0 {
		if errs := acl.NewInterpreterFromACLMap(r.ACL, false).Validate(); len(errs) > 0 {
			return fmt.Errorf("acl: %s", errs[0])
		}
	}
	if len(r.Env) > 0 {
		if errs := ValidateEnvVariables(r.Env); len(errs) > 0 {
			return fmt.Errorf("env: %s", errs[0])
		}
	}
	return nil
}

// ValidateIdentity checks whether an identity field values
// are valid.
func ValidateIdentity(i *types.Identity) error {
	if len(i.Email) == 0 {
		return fmt.Errorf("email is required")
	} else if len(i.Password) == 0 {
		return fmt.Errorf("password is required")
	} else if len(i.Password) < 8 {
		return fmt.Errorf("password is too short. Minimum of 8 characters required")
	}
	return nil
}

// ValidateEnvVariables validates the keys of a map containing environment variable data
func ValidateEnvVariables(envs map[string]string) []error {
	var errs []error
	var validFlags = []string{
		"private",
		"genRand16",
		"genRand24",
		"genRand32",
		"genRand64",
		"genRand128",
		"genRand256",
		"genRand512",
		"pin",
		"unpin",
		"unpin_once",
	}
	for k := range envs {
		if !govalidator.Matches(k, "(?i)^[a-z_0-9@,]+$") {
			errs = append(errs, fmt.Errorf("'%s': invalid key. Only alphanumeric characters and underscores are allowed", k))
		}
		if strings.Index(k, ",") != -1 && strings.Index(k, "@") == -1 {
			errs = append(errs, fmt.Errorf("'%s': invalid key. Comma (,) is only allowed when using multiple flags", k))
		}
		if strings.Index(k, "@") != -1 {
			flags := types.GetFlags(k)
			for _, f := range flags {
				if !util.InStringSlice(validFlags, f) {
					errs = append(errs, fmt.Errorf("'%s': variable has invalid flag = '%s'", k, f))
				}
			}
		}
	}
	return errs
}

// ValidateFirewallRules parses and validates the content of
// a firewall ruleset. It expects a json string or a slice
// of map[string]strings. It will return a slice of FirewallRule
// values that represents valid firewall rules. Destination addresses
// are not resolved.
func ValidateFirewallRules(firewall interface{}) ([]types.FirewallRule, []error) {

	var errs []error
	if firewall == nil {
		errs = append(errs, fmt.Errorf("function value is nil"))
		return nil, errs
	}

	var firewallMap []map[string]string

	// Coerce firewall value from either json encoded firewal to
	// []map[string]string
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
		errs = append(errs, fmt.Errorf("invalid type. expects a json string or a slice of map"))
		return nil, errs
	}

	var firewallRules []types.FirewallRule

	for i, rule := range firewallMap {
		if rule["destination"] == "" {
			errs = append(errs, fmt.Errorf("rule %d: destination is required", i))
		} else if !govalidator.IsHost(rule["destination"]) {
			errs = append(errs, fmt.Errorf("rule %d: destination is not a valid IP or host", i))
		}

		port := rule["port"]
		if port == "" {
			port = rule["destinationPort"]
		}
		if port == "" {
			errs = append(errs, fmt.Errorf("rule %d: port is required", i))
		}
		if rule["protocol"] == "" {
			rule["protocol"] = "tcp"
		} else if rule["protocol"] != "tcp" && rule["protocol"] != "udp" {
			errs = append(errs, fmt.Errorf("rule %d: invalid protocol", i))
		}

		firewallRules = append(firewallRules, types.FirewallRule{
			Destination:     rule["destination"],
			DestinationPort: port,
			Protocol:        rule["protocol"],
		})
	}

	return firewallRules, errs
}
