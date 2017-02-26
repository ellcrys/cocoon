package cmd

import (
	"github.com/ncodes/cocoon/core/server"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the cocoon engine",
	Long:  `Start the cocoon engine`,
	Run: func(cmd *cobra.Command, args []string) {

		var log = logging.MustGetLogger("start")
		var done = make(chan bool, 1)

		log.Info("Starting Cocoon core")
		server := server.NewServer()
		port, _ := cmd.Flags().GetString("port")
		go server.Start(port)

		<-done
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("port", "p", "5400", "The port to run core on")
}
