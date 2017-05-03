package cmd

import (
	"fmt"

	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login [OPTIONS] EMAIL",
	Short: "Login with your Ellcrys identity to create and manage your resources.",
	Long:  `Login with your Ellcrys identity to create and manage cocoons, vote releases and more.`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys login" requires at least 1 argument(s)`, `ellcrys login --help`)
		}

		fmt.Printf("Please enter your password: ")
		password, err := terminal.ReadPassword(0)
		if err != nil {
			log.Fatal("Err: Failed to get password")
		}

		fmt.Println("")

		if err = client.Login(args[0], string(password)); err != nil {
			log.Fatalf("Err: %s", (common.GetRPCErrDesc(err)))
		}
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
