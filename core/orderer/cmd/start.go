package cmd

import (
	"github.com/ellcrys/util"
	b_impl "github.com/ncodes/cocoon/core/blockchain/impl"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/store/impl"
	"github.com/spf13/cobra"
)

// ordererCmd represents the orderer command
var ordererCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the orderer",
	Long:  `Starts the orderer`,
	Run: func(cmd *cobra.Command, args []string) {

		var log = config.MakeLogger("orderer", "orderer")
		log.Info("Orderer has started")
		bindAddr := scheduler.Getenv("ADDR_ORDERER_RPC", "127.0.0.1:8001")
		storeConStr := util.Env("STORE_CON_STR", "host=localhost user=ned dbname=cocoon sslmode=disable password=")

		endedCh := make(chan bool)
		newOrderer := orderer.NewOrderer()
		newOrderer.SetStore(new(impl.PostgresStore))
		newOrderer.SetBlockchain(new(b_impl.PostgresBlockchain))
		go newOrderer.Start(bindAddr, storeConStr, endedCh)

		<-endedCh
	},
}

func init() {
	RootCmd.AddCommand(ordererCmd)
}
