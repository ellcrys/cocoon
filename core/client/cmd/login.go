package cmd

import (
	"github.com/ncodes/cocoon/core/client/auth"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login [email]",
	Short: "Login to authorize terminal operations",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			log.Fatal("Email address is required")
		}

		log.Info("Please enter your password:")
		password, err := terminal.ReadPassword(0)
		if err != nil {
			log.Fatal("failed to get password")
		}

		if err = auth.Login(args[0], string(password)); err != nil {
			log.Fatalf("%s", common.GetRPCErrDesc(err))
		}
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
