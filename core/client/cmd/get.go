package cmd

import (
	"github.com/ellcrys/cocoon/core/client/client"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [OPTIONS] COCOON [COCOON...]",
	Short: "Get one or more cocoons by their unique ID",
	Long:  `Get one or more cocoons by their unique ID`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys get" requires at least 1 argument(s)`, `ellcrys get --help`)
		}

		if err := client.GetCocoons(args); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}
