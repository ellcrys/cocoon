package api

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/util"
	cocoon_util "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/types"
)

var (
	// supportedLanguages list the languages supported
	supportedLanguages = []string{"go"}

	// supportedMemory list the memory values supported
	supportedMemory = []string{"512m", "1g", "2g"}

	// supportedCPUShare list the cpu share values supported
	supportedCPUShare = []string{"1x", "2x"}
)

// ValidateCocoon validates a cocoon to be created
func ValidateCocoon(c *types.Cocoon) error {

	if len(c.ID) == 0 {
		return fmt.Errorf("id is required")
	} else if !govalidator.IsUUIDv4(c.ID) {
		return fmt.Errorf("id is not a valid uuid")
	} else if len(c.URL) == 0 {
		return fmt.Errorf("url is required")
	} else if !cocoon_util.IsGithubRepoURL(c.URL) {
		return fmt.Errorf("url is not a valid github repo url")
	} else if len(c.Language) == 0 {
		return fmt.Errorf("language is required")
	} else if !util.InStringSlice(supportedLanguages, c.Language) {
		return fmt.Errorf("language is not supported. Expects one of these values %s", supportedLanguages)
	} else if len(c.BuildParam) > 0 {
		var _c map[string]interface{}
		if util.FromJSON([]byte(c.BuildParam), &_c) != nil {
			return fmt.Errorf("build parameter is not valid json")
		}
	} else if len(c.Memory) == 0 {
		return fmt.Errorf("memory is required")
	} else if !util.InStringSlice(supportedMemory, c.Memory) {
		return fmt.Errorf("Memory value is not supported. Expects one of these values %s", supportedMemory)
	} else if len(c.CPUShare) == 0 {
		return fmt.Errorf("CPU share is required")
	} else if !util.InStringSlice(supportedCPUShare, c.CPUShare) {
		return fmt.Errorf("CPU share value is not supported. Expects one of these values %s", supportedCPUShare)
	} else if c.Instances > 10 {
		return fmt.Errorf("Instances value is currently limited to 10")
	}

	return nil
}

// ValidateRelease checks whether a release field values
// are valid.
func ValidateRelease(r *types.Release) error {

	if len(r.ID) == 0 {
		return fmt.Errorf("id is required")
	} else if !govalidator.IsUUIDv4(r.ID) {
		return fmt.Errorf("id is not a valid uuid")
	} else if len(r.CocoonID) == 0 {
		return fmt.Errorf("cocoon id is required")
	} else if len(r.URL) == 0 {
		return fmt.Errorf("url is required")
	} else if !cocoon_util.IsGithubRepoURL(r.URL) {
		return fmt.Errorf("url is not a valid github repo url")
	} else if len(r.Language) == 0 {
		return fmt.Errorf("language is required")
	} else if !util.InStringSlice(supportedLanguages, r.Language) {
		return fmt.Errorf("language is not supported. Expects one of these values %s", supportedLanguages)
	} else if len(r.BuildParam) > 0 {
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
