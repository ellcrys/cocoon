package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [OPTIONS] COCOON [COCOON...]",
	Short: "Start one or more new or stopped cocoons",
	Long:  `Start one or more new or stopped cocoons`,
	Run: func(cmd *cobra.Command, args []string) {

		useLastDeployedReleaseID, _ := cmd.Flags().GetBool("last")
		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys start" requires at least 1 argument(s)`, `ellcrys start --help`)
		}

		if err := client.Start(args, useLastDeployedReleaseID); err != nil {
			desc := common.GetRPCErrDesc(err)
			switch desc {
			case "unknown service proto.API":
				desc = "unable to connect to the cluster"
			}
			log.Fatalf("Err: %s", desc)
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().BoolP("last", "", false, "Forces the execution of the last deployed release")
}
