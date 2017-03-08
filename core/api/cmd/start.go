package cmd

import (
	"log"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/grpc"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/spf13/cobra"
)

// SchedulerAddr represents the scheduler address
var SchedulerAddr = util.Env("SCHEDULER_ADDR", "")

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the API server.",
	Long:  `This command starts the API server.`,
	Run: func(cmd *cobra.Command, args []string) {
		bindAddr, _ := cmd.Flags().GetString("bind-addr")
		schedulerAddr, _ := cmd.Flags().GetString("scheduler-addr")

		// use env vars if flag not set
		if len(schedulerAddr) == 0 {
			schedulerAddr = SchedulerAddr
		}

		if len(schedulerAddr) == 0 {
			log.Fatal("scheduler address not set in flag or environment variable")
		}

		nomad := scheduler.NewNomad()
		nomad.SetAddr(schedulerAddr, false)
		grpc.SetScheduler(nomad)

		var endedCh = make(chan bool)
		api := grpc.NewAPI()
		api.Start(bindAddr, endedCh)
		<-endedCh
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("bind-addr", "", ":8004", "The address to bind to. Expect format ip:port.")
	startCmd.Flags().StringP("scheduler-addr", "", "", "The address to the scheduler")
	startCmd.Flags().BoolP("scheduler-addr-https", "", true, "Whether to use https when accessing the scheduler")
}
