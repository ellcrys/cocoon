package cmd

import (
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/blockchain/chain"
	"github.com/ncodes/cocoon/core/blockchain/orderer"
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
		port := util.Env("ORDERER_PORT", "9000")

		// start the orderer
		endedCh := make(chan bool)

		newOrderer := orderer.NewOrderer()
		newOrderer.SetChain(new(chain.PostgresChain))
		go newOrderer.Start(port, endedCh)

		<-endedCh
	},
}

func init() {
	RootCmd.AddCommand(ordererCmd)
}
