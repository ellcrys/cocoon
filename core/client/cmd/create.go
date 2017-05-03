package cmd

import (
	"io/ioutil"
	"os"

	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/util"
	"github.com/goware/urlx"
	"github.com/hashicorp/hcl"
	"github.com/jinzhu/copier"
	c_util "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/server/acl"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// parseContract passes a contract files
func parseContract(path, repoVersion string) ([]*proto_api.CocoonReleasePayloadRequest, []error) {
	var id string
	var url string
	var lang string
	var version string
	var buildParams string
	var link string
	var resourceSet = "s1"
	var selectedResourceSet map[string]int
	var numSig = 1
	var sigThreshold = 1
	var firewallRules string
	var configFileData map[string]interface{}
	var aclMap map[string]interface{}
	var cocoons []*proto_api.CocoonReleasePayloadRequest
	var env map[string]interface{}
	var errs []error
	var enableFirewall = true

	// path is a local file path
	if ok, _ := govalidator.IsFilePath(path); ok {
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
			path = ""
			break
		}
	}

	// path is a url, download it
	if govalidator.IsURL(path) {
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

	if len(configFileData) > 0 && configFileData["contracts"] != nil {
		if contracts, ok := configFileData["contracts"].([]map[string]interface{}); ok && len(contracts) > 0 {
			for i, contract := range contracts {

				id = util.UUID4()
				if _id, ok := contract["id"].(string); ok && len(_id) > 0 {
					id = _id
				}

				if repos, ok := contract["repo"].([]map[string]interface{}); ok && len(repos) > 0 {
					url = toStringOr(repos[0]["url"], "")
					version = toStringOr(repos[0]["version"], "")
					lang = toStringOr(repos[0]["language"], "")
					link = toStringOr(repos[0]["link"], "")
				} else {
					errs = append(errs, fmt.Errorf("contract %d: missing repo stanza", i))
					return nil, errs
				}

				if builds, ok := contract["build"].([]map[string]interface{}); ok && len(builds) > 0 {
					buildJSON, _ := util.ToJSON(builds[0])
					buildParams = string(buildJSON)
				}

				if resources, ok := contract["resources"].([]map[string]interface{}); ok && len(resources) > 0 {
					resourceSet = toStringOr(resources[0]["resource_set"], resourceSet)
					valid := false
					for k, v := range common.SupportedResourceSets {
						if k == resourceSet {
							valid = true
							selectedResourceSet = v
						}
					}
					if !valid {
						errs = append(errs, fmt.Errorf("resources: unknown resource_set value: %s", resourceSet))
					}
				}

				if signatories, ok := contract["signatories"].([]map[string]interface{}); ok && len(signatories) > 0 {
					numSig = toIntOr(signatories[0]["max"], numSig)
					sigThreshold = toIntOr(signatories[0]["threshold"], sigThreshold)
				}

				if acls, ok := contract["acl"].([]map[string]interface{}); ok && len(acls) > 0 {
					aclMap = acls[0]
				}

				if firewall, ok := contract["firewall"].([]map[string]interface{}); ok && len(firewall) > 0 {

					if firewall[0]["enabled"] != nil {
						enableFirewall = firewall[0]["enabled"].(bool)
					}

					if enableFirewall {
						if len(firewall[0]["rule"].([]map[string]interface{})) > 0 {
							bs, _ := util.ToJSON(firewall[0]["rule"])
							firewallRules = string(bs)
						}
					}
				}

				if envs, ok := contract["env"].([]map[string]interface{}); ok && len(envs) > 0 {
					env = envs[0]
				}

				// validate ACLMap
				if len(aclMap) > 0 {
					var _errs = acl.NewInterpreter(aclMap, false).Validate()
					if len(errs) > 0 {
						for _, err := range _errs {
							errs = append(errs, fmt.Errorf("acl: %s", err))
						}
						return nil, errs
					}
				}

				// parse and validate firewall
				var payloadFirewallRules []*proto_api.FirewallRule
				if len(firewallRules) > 0 {
					var _errs []error
					validFirewallRules, _errs := api.ValidateFirewallRules(firewallRules)
					if len(_errs) > 0 {
						for _, err := range errs {
							errs = append(errs, fmt.Errorf("firewall: %s", err))
						}
						return nil, errs
					}
					copier.Copy(&payloadFirewallRules, validFirewallRules)
				}

				cocoons = append(cocoons, &proto_api.CocoonReleasePayloadRequest{
					CocoonID:       id,
					URL:            url,
					Language:       lang,
					Version:        version,
					BuildParam:     buildParams,
					EnableFirewall: enableFirewall,
					Firewall:       payloadFirewallRules,
					ACL:            types.NewACLMap(aclMap).ToJSON(),
					Env:            types.NewEnv(env),
					Memory:         int32(selectedResourceSet["memory"]),
					CPUShare:       int32(selectedResourceSet["cpuShare"]),
					Link:           link,
					NumSignatories: int32(numSig),
					SigThreshold:   int32(sigThreshold),
				})
			}
		}
		return cocoons, nil
	}

	return nil, []error{fmt.Errorf("Unrecognised path: %s", path)}
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [OPTIONS] CONTRACT_FILE_PATH",
	Short: "Create a new cocoon",
	Long:  `Create a new cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		v, _ := cmd.Flags().GetString("version")

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys create" requires at least 1 argument(s)`, `ellcrys create --help`)
		}

		stopSpinner := util.Spinner("Please wait...")

		cocoons, errs := parseContract(args[0], v)
		if errs != nil && len(errs) > 0 {
			stopSpinner()
			for _, err := range errs {
				log.Errorf("Err: %s", err.Error())
			}
			os.Exit(1)
		}

		stopSpinner()
		for i, cocoon := range cocoons {
			err := client.CreateCocoon(cocoon)
			if err != nil {
				log.Fatalf("Err (Contract %d): %s", i, (common.GetRPCErrDesc(err)))
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().StringP("version", "v", "master", "Set the branch name or commit hash for a github hosted contract file")
}
