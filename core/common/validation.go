package common

import (
	"fmt"

	"github.com/ellcrys/util"
	cocoon_util "github.com/ncodes/cocoon-util"
)

// ValidateDeployment validates the parameters of a cocoon code deploy request
func ValidateDeployment(url, language, buildParam string) error {
	if len(url) == 0 {
		return fmt.Errorf("url is required")
	} else if !cocoon_util.IsGithubRepoURL(url) {
		return fmt.Errorf("url is not a valid github repo url")
	} else if len(language) == 0 {
		return fmt.Errorf("language is required")
	} else if !util.InStringSlice([]string{"go"}, language) {
		return fmt.Errorf("language is not supported. Expected one of these languages [go]")
	} else if len(buildParam) > 0 {
		var c map[string]interface{}
		if util.FromJSON([]byte(buildParam), &c) != nil {
			return fmt.Errorf("build parameter is not valid json")
		}
	}
	return nil
}
