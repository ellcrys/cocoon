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
		memory, _ := cmd.Flags().GetString("memory")
		cpuShare, _ := cmd.Flags().GetString("cpu-share")

		err := new(cocoon.Deploy).Deploy(url, releaseTag, lang, buildParams, memory, cpuShare)
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
	cocoonDeployCmd.Flags().StringP("memory", "m", "512m", "The amount of memory to allocate. e.g 512m, 1g or 2g")
	cocoonDeployCmd.Flags().StringP("cpu-share", "c", "1x", "The share of cpu to allocate. e.g 1x or 2x")
}
