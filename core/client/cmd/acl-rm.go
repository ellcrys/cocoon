package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// aclRmCmd represents the acl-rm command
var aclRmCmd = &cobra.Command{
	Use:   "acl-rm [OPTIONS] COCOON TARGET",
	Short: "Remove an ACL rule from a cocoon",
	Long:  `Remove an ACL rule from a cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) < 2 {
			UsageError(log, cmd, `"ellcrys acl-rm" requires at least 1 argument(s)`, `ellcrys acl-rm --help`)
		}

		if err := client.RemoveACLRule(args[0], args[1]); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(aclRmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// aclRmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// aclRmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
