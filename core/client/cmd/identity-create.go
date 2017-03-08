package cmd

import (
	"github.com/ncodes/cocoon/core/client/identity"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// identityCreateCmd represents the identity-create command
var identityCreateCmd = &cobra.Command{
	Use:   "identity-create",
	Short: "Create an identity",
	Long:  `Create an identity to use for platform operations`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		identity := identity.NewIdentity()
		pubKey, _ := cmd.Flags().GetString("pub-key")
		outputFile, _ := cmd.Flags().GetString("output")

		if len(args) == 0 {
			log.Fatal("Email is required")
		}

		if err := identity.Create(args[0], pubKey, outputFile); err != nil {
			log.Fatalf("%s", common.StripRPCErrorPrefix([]byte(err.Error())))
		}
	},
}

func init() {
	RootCmd.AddCommand(identityCreateCmd)
	identityCreateCmd.PersistentFlags().String("pub-key", "", "An ECDSA public key for verifying operations")
	identityCreateCmd.PersistentFlags().StringP("output", "o", "", "Write key to a file instead of stdout")
}
