package cmd

import (
	"fmt"

	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [OPTIONS] COCOON [COCOON...]",
	Short: "Get a cocoon by its unique ID",
	Long:  `Get a cocoon by its unique ID`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys get" requires at least 1 argument(s)`, `ellcrys get --help`)
		}
		// TODO: Work your own magic here
		fmt.Println("get called")
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}
