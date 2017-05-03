package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// identityCreateCmd represents the identity-create command
var identityCreateCmd = &cobra.Command{
	Use:   "new-id [OPTIONS] EMAIL",
	Short: "Create an Ellcrys platform identity",
	Long:  `Create an Ellcrys platform identity`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys new-id" requires at least 1 argument(s)`, `ellcrys new-id --help`)
		}

		if err := client.CreateIdentity(args[0]); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(identityCreateCmd)
}
