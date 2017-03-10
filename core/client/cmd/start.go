package cmd

import (
	"github.com/ncodes/cocoon/core/client/cocoon"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a new or stopped cocoon",
	Long:  `Starts a new or stopped cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			log.Fatal("Cocoon ID is required")
		}

		ops := new(cocoon.Ops)
		if err := ops.Start(args[0]); err != nil {
			log.Fatalf("%s", common.StripRPCErrorPrefix([]byte(err.Error())))
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}