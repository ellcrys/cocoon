package cmd

import (
	"os"
	"time"

	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/scheduler"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

var apiLog *logging.Logger

var (
	bindAddr = util.Env("BIND_ADDR", "")
)

func init() {
	config.ConfigureLogger()
	apiLog = config.MakeLogger("api")
	goreq.SetConnectTimeout(5 * time.Second)
}

// apiStartCmd
var apiStartCmd = &cobra.Command{
	Use:   "start [OPTIONS]",
	Short: "Start the Platform API server",
	Long:  "Start the Platform API server",
	Run: func(cmd *cobra.Command, args []string) {

		// Ensure expected environment variables are set
		if missingEnv := common.HasEnv([]string{
			"ENV",
			"API_SIGN_KEY",
			"API_VERSION",
			"CONNECTOR_VERSION",
			"GCP_PROJECT_ID",
			"REPO_ARCHIVE_BKT",
		}...); len(missingEnv) > 0 {
			apiLog.Fatalf("The following environment variables must be set: %v", missingEnv)
		}

		// set bind address from environment set by scheduler
		if len(bindAddr) == 0 {
			bindAddr = scheduler.Getenv("ADDR_API_RPC", "127.0.0.1:8005")
		}

		api, err := api.NewAPI()
		if err != nil {
			apiLog.Fatal(err.Error())
		}

		common.OnTerminate(func(s os.Signal) {
			apiLog.Info("Terminate signal received. Stopping...")
			api.Stop()
		})

		var endedCh = make(chan bool)
		api.Start(bindAddr, endedCh)
		<-endedCh
		apiLog.Info("Stopped")
	},
}

func init() {
	RootCmd.AddCommand(apiStartCmd)
	apiStartCmd.Flags().BoolP("scheduler-addr-protocol", "", true, "Whether to use https or http when accessing the scheduler")
}
