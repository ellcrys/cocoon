package cmd

import (
	"os"

	"github.com/ellcrys/util"
	"github.com/ellcrys/cocoon/core/client/client"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

func isValidResourceSet(resourceSet string) (map[string]int, bool) {
	for k, v := range common.SupportedResourceSets {
		if k == resourceSet {
			return v, true
		}
	}
	return nil, false
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [OPTIONS] CONTRACT_FILE_PATH",
	Short: "Create a new cocoon",
	Long:  `Create a new cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		repoVersion, _ := cmd.Flags().GetString("repo-version")

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys create" requires at least 1 argument(s)`, `ellcrys create --help`)
		}

		stopSpinner := util.Spinner("Please wait")

		contracts, errs := makeContractFromFile(args[0], repoVersion)
		if errs != nil && len(errs) > 0 {
			stopSpinner()
			for _, err := range errs {
				log.Errorf("Err: %s", err.Error())
			}
			os.Exit(1)
		}

		stopSpinner()
		for i, contract := range contracts {

			// modify fields if corresponding flag is set
			errs := modifyFromFlags(cmd, contract)
			if len(errs) > 0 {
				for _, err := range errs {
					log.Errorf("Err (Contract %d): %s", i, (common.GetRPCErrDesc(err)))
				}
				continue
			}

			err := client.CreateCocoon(contract)
			if err != nil {
				log.Fatalf("Err (Contract %d): %s", i, (common.GetRPCErrDesc(err)))
			}
		}
	},
}

// addContractModifierFlags adds flags that can be used to modify a
// contract. Set forUpdate to true to not include flags for fields that are immutable
func addContractModifierFlags(cmd *cobra.Command, forUpdate bool) {
	if !forUpdate {
		cmd.Flags().String("id", "", "The id of the contract")
	}
	cmd.Flags().String("repo.url", "", "The github repository URL")
	cmd.Flags().String("repo.version", "", "The github repository release tag or commit hash")
	cmd.Flags().String("repo.language", "", "The language used to develop the cocoon code")
	cmd.Flags().String("repo.link", "", "An optional cocoon id to link to")
	cmd.Flags().String("build.pkgMgr", "", "The package manage to use during build stage")
	cmd.Flags().String("resources.set", "", "The resource set to apply to the cocoon")
	cmd.Flags().Int("signatories.max", -1, "The maximum number of signatories required")
	cmd.Flags().Int("signatories.threshold", -1, "The minimum required signatures/votes to approve a release")
	cmd.Flags().String("acl", "", "Access control list")
	cmd.Flags().String("firewall.enabled", "", "Enable firewall rules")
	cmd.Flags().String("firewall.rules", "", "Firewall rules to attach to the cocoon")
	cmd.Flags().String("env", "", "Environment variables to pass the cocoon")
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().String("repo-version", "master", "Set the branch name or commit hash for a github hosted contract file")
	addContractModifierFlags(createCmd, false)
}
