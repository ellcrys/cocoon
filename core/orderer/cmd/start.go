package cmd

import (
	"log"
	"time"

	"os"

	"github.com/ellcrys/util"
	b_impl "github.com/ellcrys/cocoon/core/blockchain/impl"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	"github.com/ellcrys/cocoon/core/lock/consul"
	"github.com/ellcrys/cocoon/core/lock/memory"
	"github.com/ellcrys/cocoon/core/orderer/orderer"
	"github.com/ellcrys/cocoon/core/scheduler"
	"github.com/ellcrys/cocoon/core/store/impl"
	"github.com/spf13/cobra"
)

func init() {
	consul.LockTTL = time.Second * 15
	memory.LockTTL = time.Second * 15
}

// ordererCmd represents the orderer command
var ordererCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the orderer",
	Long:  `Starts the orderer`,
	Run: func(cmd *cobra.Command, args []string) {

		// Ensure expected environment variables are set
		if missingEnv := common.HasEnv([]string{}...); len(missingEnv) > 0 {
			log.Fatalf("The following environment variables must be set: %v", missingEnv)
		}

		if os.Getenv("DEV_MEMORY_LOCK") != "" {
			c := memory.StartLockWatcher()
			defer c()
		}

		var log = config.MakeLogger("orderer")
		log.Info("Orderer has started")
		bindAddr := scheduler.Getenv("ADDR_ORDERER_RPC", "127.0.0.1:8001")
		storeConStr := util.Env("STORE_CON_STR", "host=localhost user=ned dbname=cocoon sslmode=disable password=")

		endedCh := make(chan bool)
		newOrderer := orderer.NewOrderer()
		newOrderer.SetStore(new(impl.PostgresStore))
		newOrderer.SetBlockchain(new(b_impl.PostgresBlockchain))
		go newOrderer.Start(bindAddr, storeConStr, endedCh)

		common.OnTerminate(func(s os.Signal) {
			log.Info("Terminate signal received. Stopping...")
			newOrderer.Stop()
		})

		<-endedCh
		log.Info("Stopped")
	},
}

func init() {
	RootCmd.AddCommand(ordererCmd)
}
