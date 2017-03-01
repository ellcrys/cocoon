package cmd

import (
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/ledgerchain/impl"
	"github.com/ncodes/cocoon/core/orderer"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// ordererCmd represents the orderer command
var ordererCmd = &cobra.Command{
	Use:   "orderer",
	Short: "The orderer is the gateway to the immutable data store",
	Long:  `The orderer manages interaction between the data store and the rest of the cluster.`,
	Run: func(cmd *cobra.Command, args []string) {

		var log = logging.MustGetLogger("orderer")
		log.Info("Orderer has started")
		addr := util.Env("ORDERER_ADDR", "127.0.0.1:8001")
		ledgerChainConStr := util.Env("LEDGER_CHAIN_CONNECTION_STRING", "host=localhost user=ned dbname=cocoonchain sslmode=disable password=")

		endedCh := make(chan bool)
		newOrderer := orderer.NewOrderer()
		newOrderer.SetLedgerChain(new(impl.PostgresLedgerChain))
		go newOrderer.Start(addr, ledgerChainConStr, endedCh)

		<-endedCh
	},
}

func init() {
	RootCmd.AddCommand(ordererCmd)
}
