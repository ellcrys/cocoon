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
		port := util.Env("ORDERER_PORT", "9000")
		log.Infof("Starting orderer GRPC server on port %s", port)

		// start the orderer
		startedCh := make(chan bool)
		endedCh := make(chan bool)

		newOrderer := orderer.NewOrderer()
		go newOrderer.Start(port, startedCh, endedCh)

		// set chain implementation after orderer starts
		<-startedCh
		newOrderer.SetChain(new(chain.PostgresChain))

		<-endedCh
	},
}

func init() {
	RootCmd.AddCommand(ordererCmd)
}
