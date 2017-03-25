package cmd

import (
	"os"
	"time"

	"fmt"

	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/scheduler"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

var log *logging.Logger

func init() {
	config.ConfigureLogger()
	log = logging.MustGetLogger("api")
	goreq.SetConnectTimeout(5 * time.Second)
}

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start and manage an API server.",
	Long:  `Start and manage an API server.`,
	Run:   nil,
}

// startCmd
var apiCmdStart = &cobra.Command{
	Use:   "start",
	Short: "Start the API server",
	Long:  "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {

		nomad := scheduler.NewNomad()
		bindAddr, _ := cmd.Flags().GetString("bind-addr")
		schedulerAddr, _ := cmd.Flags().GetString("scheduler-addr")

		// set scheduler addr from environment var if set
		if len(schedulerAddr) == 0 {
			schedulerAddr = os.Getenv("SCHEDULER_ADDR")
		}

		// Try to discover scheduler address
		if len(schedulerAddr) == 0 {
			services, err := nomad.ServiceDiscovery.GetByID(nomad.GetName(), map[string]string{"tag": "http"})
			if err != nil {
				log.Fatalf("failed to get scheduler service from discovery service. %s", err)
			}
			if len(services) > 0 {
				schedulerAddr = fmt.Sprintf("%s:%d", services[0].IP, int(services[0].Port))
			}
		}

		if len(schedulerAddr) == 0 {
			log.Fatal("scheduler address not set in flag or environment variable")
		}

		// set bind address from environment var set by scheduler
		if len(bindAddr) == 0 {
			bindAddr = scheduler.Getenv("ADDR_API_RPC", "127.0.0.1:8005")
		}

		nomad.SetAddr(schedulerAddr, false)
		api := api.NewAPI(nomad)

		var endedCh = make(chan bool)
		api.Start(bindAddr, endedCh)
		<-endedCh
	},
}

func init() {
	apiCmd.AddCommand(apiCmdStart)
	RootCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringP("bind-addr", "", ":8005", "The address to bind to. Expect format ip:port.")
	apiCmd.Flags().StringP("scheduler-addr", "", "", "The address to the scheduler")
	apiCmd.Flags().BoolP("scheduler-addr-protocol", "", true, "Whether to use https or http when accessing the scheduler")
}
