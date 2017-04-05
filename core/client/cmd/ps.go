package cmd

import (
	"log"

	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
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
		if err := client.ListCocoons(showAll, jsonFormatted); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(psCmd)
	psCmd.PersistentFlags().BoolP("all", "a", false, "Show all cocoons (default shows just running)")
	psCmd.PersistentFlags().BoolP("json", "", false, "Return result as JSON formatted output")
}
