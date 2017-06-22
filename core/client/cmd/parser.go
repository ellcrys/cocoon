package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/util"
	"github.com/goware/urlx"
	"github.com/hashicorp/hcl"
	"github.com/jinzhu/copier"
	c_util "github.com/ncodes/cocoon-util"
	"github.com/ellcrys/cocoon/core/api/api"
	"github.com/ellcrys/cocoon/core/api/api/proto_api"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/connector/server/acl"
	"github.com/ellcrys/cocoon/core/types"
	"github.com/spf13/cobra"
)

// makeContractFromFile fetches a contract file from
// local file system, public github repository or a remote server and
// parses it, returning a slice of contracts
func makeContractFromFile(path, repoVersion string) ([]*proto_api.ContractRequest, []error) {

	var configFileData map[string]interface{}
	var cocoons []*proto_api.ContractRequest
	var errs []error

	// path is a github url, download contract from the root of the master branch
	if govalidator.IsURL(path) && c_util.IsGithubRepoURL(path) {
		url, _ := urlx.Parse(path)
		urls := []string{
			fmt.Sprintf("https://raw.githubusercontent.com%s/%s/contract.hcl", url.Path, repoVersion),
			fmt.Sprintf("https://raw.githubusercontent.com%s/%s/contract.json", url.Path, repoVersion),
		}
		for _, url := range urls {
			var fileData []byte
			err := util.DownloadURLToFunc(url, func(b []byte, code int) error {
				if code == 404 {
					return fmt.Errorf("contract file not found")
				}
				fileData = append(fileData, b...)
				if len(fileData) > 5000000 {
					return fmt.Errorf("Maximum contract file size reached. aborting download")
				}
				return nil
			})
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to download contract file: %s", err))
				return nil, errs
			}
			if err = hcl.Decode(&configFileData, string(fileData)); err != nil {
				errs = append(errs, fmt.Errorf("failed to parse contract file: %s", err.Error()))
				return nil, errs
			}
			break
		}
	}

	// path is a url, download it
	if len(configFileData) == 0 && govalidator.IsURL(path) {
		var fileData []byte
		err := util.DownloadURLToFunc(path, func(b []byte, code int) error {
			if code == 404 {
				return fmt.Errorf("contract file not found")
			}
			fileData = append(fileData, b...)
			if len(fileData) > 5000000 {
				return fmt.Errorf("Maximum contract file size reached. aborting download")
			}
			return nil
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to download contract file: %s", err))
			return nil, errs
		}
		if err = hcl.Decode(&configFileData, string(fileData)); err != nil {
			errs = append(errs, fmt.Errorf("failed to parse contract file: %s", err.Error()))
			return nil, errs
		}
	}

	// if we got here without the configFileData having any value, assume the path is a local file path
	if len(configFileData) == 0 {
		path, _ := filepath.Abs(path)
		fileData, err := ioutil.ReadFile(path)
		if err != nil {
			errs = append(errs, err)
			return nil, errs
		}
		if err = hcl.Decode(&configFileData, string(fileData)); err != nil {
			errs = append(errs, fmt.Errorf("failed to parse contract file: %s", err.Error()))
			return nil, errs
		}
	}

	if len(configFileData) == 0 || configFileData["contracts"] == nil {
		return nil, []error{fmt.Errorf("Unrecognised path: %s", path)}
	}

	// parse each contracts contained in the config file data
	if contracts, ok := configFileData["contracts"].([]map[string]interface{}); ok {

		var contract proto_api.ContractRequest

		for i, _contract := range contracts {

			// parse 'id'
			if id, ok := _contract["id"].(string); ok && len(id) > 0 {
				contract.CocoonID = id
			} else {
				contract.CocoonID = util.UUID4()
			}

			// parse 'repo' stanza
			if repos, ok := _contract["repo"].([]map[string]interface{}); ok && len(repos) > 0 {
				contract.URL = toStringOr(repos[0]["url"], "")
				contract.Version = toStringOr(repos[0]["version"], "")
				contract.Language = toStringOr(repos[0]["language"], "")
				contract.Link = toStringOr(repos[0]["link"], "")
			} else {
				errs = append(errs, fmt.Errorf("contract %d: missing repo stanza", i))
				return nil, errs
			}

			// parse 'build' stanza
			if builds, ok := _contract["build"].([]map[string]interface{}); ok && len(builds) > 0 {
				mergedBuild := common.MergeMapSlice(builds)
				buildJSON, _ := util.ToJSON(mergedBuild)
				contract.BuildParam = string(buildJSON)
			}

			// parse 'resources' stanza
			if resources, ok := _contract["resources"].([]map[string]interface{}); ok && len(resources) > 0 {
				mergedResources := common.MergeMapSlice(resources)
				resourceSet := toStringOr(mergedResources["set"], "s1")
				if selectedResourceSet, valid := isValidResourceSet(resourceSet); valid {
					contract.Memory = int32(selectedResourceSet["memory"])
					contract.CPUShare = int32(selectedResourceSet["cpuShare"])
				} else {
					errs = append(errs, fmt.Errorf("resources: unknown set value: %s", resourceSet))
				}
			}

			// parse 'signatories' stanza
			if signatories, ok := _contract["signatories"].([]map[string]interface{}); ok && len(signatories) > 0 {
				mergedSignatories := common.MergeMapSlice(signatories)
				contract.NumSignatories = int32(toIntOr(mergedSignatories["max"], 1))
				contract.SigThreshold = int32(toIntOr(mergedSignatories["threshold"], 1))
			}

			// parse 'acl' stanza
			if acls, ok := _contract["acl"].([]map[string]interface{}); ok && len(acls) > 0 {
				mergedACLs := common.MergeMapSlice(acls)
				if _errs := acl.NewInterpreter(mergedACLs, false).Validate(); len(_errs) > 0 {
					for _, err := range _errs {
						errs = append(errs, fmt.Errorf("acl: %s", err))
					}
					return nil, errs
				}
				contract.ACL = types.NewACLMap(mergedACLs).ToJSON()
			}

			// parse 'firewall' stanza
			if firewall, ok := _contract["firewall"].([]map[string]interface{}); ok && len(firewall) > 0 {
				mergedFirewall := common.MergeMapSlice(firewall)
				if enabledFirewall, ok := mergedFirewall["enabled"].(bool); ok {
					contract.EnableFirewall = enabledFirewall
					if !contract.EnableFirewall {
						contract.Firewall = nil
					} else {
						if rules, ok := mergedFirewall["rule"].([]map[string]interface{}); ok && len(rules) > 0 {
							bs, _ := util.ToJSON(rules)
							validFirewallRules, _errs := api.ValidateFirewallRules(string(bs))
							if len(_errs) > 0 {
								for _, err := range _errs {
									errs = append(errs, fmt.Errorf("firewall: %s", err))
								}
								return nil, errs
							}
							copier.Copy(&contract.Firewall, validFirewallRules)
						}
					}
				}
			}

			// parse 'env' stanza
			if envs, ok := _contract["env"].([]map[string]interface{}); ok && len(envs) > 0 {
				contract.Env = types.NewEnv(common.MergeMapSlice(envs))
			}

			cocoons = append(cocoons, &contract)
		}
	}
	return cocoons, nil
}

// modifyFromFlags takes a contract and updates the field with
// values from corresponding flags
func modifyFromFlags(cmd *cobra.Command, contract *proto_api.ContractRequest) []error {

	var errs []error

	// Repository ID
	if id, _ := cmd.Flags().GetString("id"); id != "" {
		contract.CocoonID = id
	}

	// Repository
	if url, _ := cmd.Flags().GetString("repo.url"); url != "" {
		contract.URL = url
	}

	// Repository version
	if version, _ := cmd.Flags().GetString("repo.version"); version != "" {
		contract.Version = version
	}

	// Repository language
	if language, _ := cmd.Flags().GetString("repo.language"); language != "" {
		contract.Language = language
	}

	// Set repository link
	if link, _ := cmd.Flags().GetString("repo.link"); link != "" {
		contract.Link = link
	}

	// build
	if buildPkgMgr, _ := cmd.Flags().GetString("build.pkgMgr"); buildPkgMgr != "" {
		var existingBuildParams map[string]interface{}
		util.FromJSON([]byte(contract.BuildParam), &existingBuildParams)
		existingBuildParams["pkgMgr"] = buildPkgMgr
		buildPkgMgr, _ := util.ToJSON(existingBuildParams)
		contract.BuildParam = string(buildPkgMgr)
	}

	// Set resource
	if resourceSet, _ := cmd.Flags().GetString("resources.set"); resourceSet != "" {
		if selectedResourceSet, valid := isValidResourceSet(resourceSet); valid {
			contract.Memory = int32(selectedResourceSet["memory"])
			contract.CPUShare = int32(selectedResourceSet["cpuShare"])
		} else {
			errs = append(errs, fmt.Errorf("resources: unknown set value: %s", resourceSet))
		}
	}

	if numSignatories, _ := cmd.Flags().GetInt("signatories.max"); numSignatories > -1 {
		contract.NumSignatories = int32(numSignatories)
	}

	if sigThreshold, _ := cmd.Flags().GetInt("signatories.threshold"); sigThreshold > -1 {
		contract.SigThreshold = int32(sigThreshold)
	}

	// parse acl in the form: '*:deny,ledger1.cocoon_id:deny'
	if aclStr, _ := cmd.Flags().GetString("acl"); len(aclStr) > 0 {
		aclSplit := strings.Split(aclStr, ",")
		aclMap := map[string]interface{}{}
		for _, aclKV := range aclSplit {
			aclKVSplit := strings.SplitN(aclKV, ":", 2)
			if !strings.Contains(aclKVSplit[0], ".") {
				aclMap[aclKVSplit[0]] = aclKVSplit[1]
				continue
			}
			aclKeySplit := strings.SplitN(aclKVSplit[0], ".", 2)
			aclMap[aclKeySplit[0]] = map[string]interface{}{
				aclKeySplit[1]: aclKeySplit[1],
			}
		}
		if _errs := acl.NewInterpreter(aclMap, false).Validate(); len(_errs) > 0 {
			for _, err := range _errs {
				errs = append(errs, fmt.Errorf("acl: %s", err))
			}
		} else {
			contract.ACL = types.NewACLMap(aclMap).ToJSON()
		}
	}

	if firewallEnabled, _ := cmd.Flags().GetString("firewall.enabled"); len(firewallEnabled) > 0 {
		contract.EnableFirewall = firewallEnabled == "true"
		if !contract.EnableFirewall {
			contract.Firewall = nil
		}
	}

	// parse firewall rules added in the form: 'destination:0.0.0.0;port:4444;protocol:tcp,destination:1.1.1.1;port:3333;protocol:udp'
	if firewallRules, _ := cmd.Flags().GetString("firewall.rules"); len(firewallRules) > 0 {
		var parsedRules []map[string]string
		rules := strings.Split(firewallRules, ",")

		for _, rule := range rules {
			ruleFields := strings.SplitN(rule, ";", 3)
			if len(ruleFields) != 3 {
				continue
			}

			newFirewallRule := map[string]string{}

			for _, field := range ruleFields {
				fieldKV := strings.SplitN(field, ":", 2)

				if fieldKV[0] == "destination" {
					newFirewallRule["destination"] = fieldKV[1]
				}
				if fieldKV[0] == "port" {
					newFirewallRule["port"] = fieldKV[1]
				}
				if fieldKV[0] == "protocol" {
					newFirewallRule["protocol"] = fieldKV[1]
				}
			}

			parsedRules = append(parsedRules, newFirewallRule)
		}

		// validate
		bs, _ := util.ToJSON(parsedRules)
		validFirewallRules, _errs := api.ValidateFirewallRules(string(bs))
		if len(_errs) > 0 {
			for _, err := range _errs {
				errs = append(errs, fmt.Errorf("firewall: %s", err))
			}
		} else {
			contract.Firewall = nil
			copier.Copy(&contract.Firewall, validFirewallRules)
		}
	}

	// parse env in the form: NAME:ben,AGE:23
	if env, _ := cmd.Flags().GetString("env"); len(env) > 0 {
		var newEnv = make(map[string]string)
		envVars := strings.Split(env, ",")
		for _, aVar := range envVars {
			kv := strings.SplitN(aVar, ":", 2)
			if len(kv) == 2 {
				newEnv[kv[0]] = kv[1]
			}
		}
		contract.Env = newEnv
	}

	return errs
}
