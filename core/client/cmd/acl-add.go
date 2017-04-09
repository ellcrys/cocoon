package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// aclAddCmd represents the acl-add command
var aclAddCmd = &cobra.Command{
	Use:   "acl-add [OPTIONS] COCOON TARGET PRIVILEGES",
	Short: "Add an ACL rule to a cocoon",
	Long:  `Add an ACL rule to a cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) < 3 {
			UsageError(log, cmd, `"ellcrys acl-add" requires at least 3 argument(s)`, `ellcrys acl-add --help`)
		}

		if err := client.AddACLRule(args[0], args[1], args[2]); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(aclAddCmd)
}
