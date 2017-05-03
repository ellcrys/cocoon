package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// sigRmCmd represents the sig-rm command
var sigRmCmd = &cobra.Command{
	Use:   "sig-rm [OPTIONS] COCOON IDENTITY [IDENTITY...]",
	Short: "Remove one or more signatories from a cocoon",
	Long:  `Remove one or more signatories from a cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) < 2 {
			UsageError(log, cmd, `"ellcrys sig-rm" requires at least 2 argument(s)`, `ellcrys sig-add --help`)
		}

		if err := client.RemoveSignatories(args[0], args[1:]); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(sigRmCmd)
}
