package cmd

import (
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/connector/server/proto"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// testConnectorCmd represents the test-connector command
var testConnectorCmd = &cobra.Command{
	Use:   "test-connector",
	Short: "Playground for testing connector during development",
	Long:  `Playground for testing connector during development`,
	Run: func(cmd *cobra.Command, args []string) {
		f, _ := cmd.Flags().GetString("func")
		var log = logging.MustGetLogger("connector-test")
		connectorAddr := ":8002"
		conn, err := grpc.Dial(connectorAddr, grpc.WithInsecure())
		defer conn.Close()
		if err != nil {
			log.Fatalf("Failed to connect to connector. Is the connector running on %s", connectorAddr)
		}

		client := proto.NewAPIClient(conn)
		resp, err := client.Invoke(context.Background(), &proto.InvokeRequest{
			Id:       util.UUID4(),
			Function: f,
			Params:   []string{"accountxxxxx"},
		})

		log.Info("Sent: ", resp, err)
	},
}

func init() {
	RootCmd.AddCommand(testConnectorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	testConnectorCmd.PersistentFlags().String("func", "f", "The function to run")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testConnectorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
