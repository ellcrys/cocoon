package cmd

import (
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// testOrderCmd represents the test-orderer command
var testOrderCmd = &cobra.Command{
	Use:   "test-orderer",
	Short: "Playground for testing orderer during development.",
	Long:  `Playground for testing orderer during development.`,
	Run: func(cmd *cobra.Command, args []string) {
		var log = logging.MustGetLogger("orderer-test")
		ordererAddr := ":8001"
		conn, err := grpc.Dial(ordererAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to orderer. Is orderer running on %s", ordererAddr)
		}

		log.Infof("Ready for test. %v", conn)
	},
}

func init() {
	RootCmd.AddCommand(testOrderCmd)
}
