package cmd

import (
	"github.com/ncodes/cocoon/core/api/grpc"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the API server.",
	Long:  `This command starts the API server.`,
	Run: func(cmd *cobra.Command, args []string) {
		bindAddr, _ := cmd.Flags().GetString("bind-addr")
		var endedCh = make(chan bool)
		api := grpc.NewAPI()
		api.Start(bindAddr, endedCh)
		<-endedCh
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("bind-addr", "", ":8004", "")
}
