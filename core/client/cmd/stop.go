package cmd

import (
	"github.com/ellcrys/cocoon/core/client/client"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [OPTIONS] COCOON [COCOON...]",
	Short: "Stop one or more running cocoons",
	Long:  `Stop one or more running cocoons`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys stop" requires at least 1 argument(s)`, `ellcrys stop --help`)
		}

		if err := client.StopCocoon(args); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(stopCmd)
}
