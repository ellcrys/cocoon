package cmd

import (
	"fmt"

	"github.com/ellcrys/util"
	b_impl "github.com/ncodes/cocoon/core/blockchain/impl"
	"github.com/ncodes/cocoon/core/orderer"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/store/impl"
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
		ip := util.Env(scheduler.Getenv("IP_orderer_grpc"), "127.0.0.1")
		port := util.Env(scheduler.Getenv("PORT_orderer_grpc"), "8001")
		storeConStr := util.Env("STORE_CON_STR", "host=localhost user=ned dbname=cocoon sslmode=disable password=")

		endedCh := make(chan bool)
		newOrderer := orderer.NewOrderer()
		newOrderer.SetStore(new(impl.PostgresStore))
		newOrderer.SetBlockchain(new(b_impl.PostgresBlockchain))
		go newOrderer.Start(fmt.Sprintf("%s:%s", ip, port), storeConStr, endedCh)

		<-endedCh
	},
}

func init() {
	RootCmd.AddCommand(ordererCmd)
}
