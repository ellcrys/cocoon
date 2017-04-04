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
	Use:   "identity-create",
	Short: "Create an identity",
	Long:  `Create an identity to use for platform operations`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			log.Fatal("Err: Email is required")
		}

		if err := client.CreateIdentity(args[0]); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(identityCreateCmd)
}
