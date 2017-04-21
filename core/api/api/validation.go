package api

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/util"
	cocoon_util "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/acl"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/types"
)

// ValidateCocoon validates a cocoon to be created
func ValidateCocoon(c *types.Cocoon) error {

	if len(c.ID) == 0 {
		return fmt.Errorf("id is required")
	}
	if !common.IsValidResName(c.ID) {
		return fmt.Errorf("id is not a valid resource name")
	}
	if len(c.URL) == 0 {
		return fmt.Errorf("url is required")
	}
	if !cocoon_util.IsGithubRepoURL(c.URL) {
		return fmt.Errorf("url is not a valid github repo url")
	}
	if !cocoon_util.IsExistingGithubRepo(c.URL) {
		return fmt.Errorf("repository url could not be reached")
	}
	if cocoon_util.IsGithubCommitID(c.Version) && !cocoon_util.IsValidGithubCommitID(c.URL, c.Version) {
		return fmt.Errorf("repository version appears to be a commit id but it does not exist")
	}
	if len(c.Language) == 0 {
		return fmt.Errorf("language is required")
	}
	if !util.InStringSlice(scheduler.SupportedCocoonCodeLang, c.Language) {
		return fmt.Errorf("language is not supported. Expects one of these values %s", scheduler.SupportedCocoonCodeLang)
	}
	if len(c.BuildParam) > 0 {
		var _c map[string]interface{}
		if util.FromJSON([]byte(c.BuildParam), &_c) != nil {
			return fmt.Errorf("build parameter is not valid json")
		}
	}
	if c.Memory == 0 {
		return fmt.Errorf("memory is required")
	}
	if c.CPUShare == 0 {
		return fmt.Errorf("CPU share is required")
	}
	if common.GetResourceSet(c.Memory, c.CPUShare) == nil {
		return fmt.Errorf("Unknown resource set")
	}
	if c.NumSignatories <= 0 {
		return fmt.Errorf("number of signatories cannot be less than 1")
	}
	if c.SigThreshold <= 0 {
		return fmt.Errorf("signatory threshold cannot be less than 1")
	}
	if c.NumSignatories < len(c.Signatories) {
		return fmt.Errorf("max signatories already added. You can't add more")
	}
	if c.Firewall != nil {
		fmt.Println("Firewall=")
		util.Printify(c.Firewall.ToMap())
		_, errs := ValidateFirewall(c.Firewall.ToMap())
		if len(errs) != 0 {
			return fmt.Errorf("firewall: %s, ", errs[0])
		}
	}
	if len(c.ACL) > 0 {
		if errs := acl.NewInterpreterFromACLMap(c.ACL, false).Validate(); len(errs) > 0 {
			return fmt.Errorf("acl: %s", errs[0])
		}
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

// ValidateFirewall parses and validates the content of
// a firewall ruleset. It expects a json string or a slice
// of map[string]strins. It will return a slice of map[string]string
// values that represents valid firewall rules. Destination host addresses
// are not resolved.
func ValidateFirewall(firewall interface{}) ([]types.FirewallRule, []error) {

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
