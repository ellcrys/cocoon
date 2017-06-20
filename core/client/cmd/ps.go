package cmd

import (
	"github.com/ellcrys/cocoon/core/client/client"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps [OPTIONS]",
	Short: "List cocoons",
	Long:  `List cocoons`,
	Run: func(cmd *cobra.Command, args []string) {
		showAll, _ := cmd.Flags().GetBool("all")
		jsonFormatted, _ := cmd.Flags().GetBool("json")

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if err := client.ListCocoons(showAll, jsonFormatted); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(psCmd)
	psCmd.PersistentFlags().BoolP("all", "a", false, "Show all cocoons (default shows just running)")
	psCmd.PersistentFlags().BoolP("json", "", false, "Return result as JSON formatted output")
}
