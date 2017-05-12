package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout [OPTIONS]",
	Short: "Log out from current session or all sessions",
	Long:  `Log out from current session or all sessions`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		allSessions, _ := cmd.Flags().GetBool("all")

		if err := client.Logout(allSessions); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
	logoutCmd.Flags().BoolP("all", "a", false, "Log out all sessions")
}
