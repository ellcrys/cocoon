package cmd

import (
	"strconv"

	"github.com/ellcrys/cocoon/core/client/client"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// sigVoteCmd represents the sig-vote command
var sigVoteCmd = &cobra.Command{
	Use:   "sig-vote [OPTIONS] RELEASE VOTE",
	Short: "Vote to approve or deny a cocoon release",
	Long:  `Vote to approve or deny a cocoon release`,
	Run: func(cmd *cobra.Command, args []string) {

		isCid, _ := cmd.Flags().GetBool("cid")
		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) <= 1 {
			UsageError(log, cmd, `"ellcrys sig-vote" requires at least 2 argument(s)`, `ellcrys sig-vote --help`)
		} else if args[1] != "0" && args[1] != "1" {
			log.Fatal("Err: Vote value must be either 1 (Approve) or 0 (Deny)")
		}

		vote, _ := strconv.Atoi(args[1])
		if err := client.AddVote(args[0], vote, isCid); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(sigVoteCmd)
	sigVoteCmd.PersistentFlags().BoolP("cid", "", false, "Force release ID to be interpreted as a cocoon ID (Default: false)")
}
