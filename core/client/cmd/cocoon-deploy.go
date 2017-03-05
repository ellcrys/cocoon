package cmd

import (
	"github.com/ncodes/cocoon/core/client/cocoon"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// cocoonDeployCmd represents the cocoon-deploy command
var cocoonDeployCmd = &cobra.Command{
	Use:   "cocoon-deploy",
	Short: "Launch a cocoon code",
	Long:  `Start a cocoon code from a github repository`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		url, _ := cmd.Flags().GetString("url")
		lang, _ := cmd.Flags().GetString("lang")
		releaseTag, _ := cmd.Flags().GetString("release-tag")
		buildParams, _ := cmd.Flags().GetString("build-param")

		err := new(cocoon.Deploy).Deploy(url, releaseTag, lang, buildParams)
		if err != nil {
			log.Fatalf("%s", common.StripRPCErrorPrefix([]byte(err.Error())))
		}
	},
}

func init() {
	RootCmd.AddCommand(cocoonDeployCmd)
	cocoonDeployCmd.Flags().StringP("url", "u", "", "Deploys a cocoon code to the platform")
	cocoonDeployCmd.Flags().StringP("lang", "l", "", "The langauges the cocoon code is written in.")
	cocoonDeployCmd.Flags().StringP("release-tag", "", "", "The github release tag. Defaults to `latest`.")
	cocoonDeployCmd.Flags().StringP("build-param", "", "", "Build parameters to apply during cocoon code build process")
}
