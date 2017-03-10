package cmd

import (
	"github.com/ncodes/cocoon/core/client/cocoon"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/types/client"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cocoon configuration locally",
	Long:  `Create or construct a cocoon locally before deploying to the live platform`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		url, _ := cmd.Flags().GetString("url")
		lang, _ := cmd.Flags().GetString("lang")
		releaseTag, _ := cmd.Flags().GetString("release-tag")
		buildParams, _ := cmd.Flags().GetString("build-param")
		memory, _ := cmd.Flags().GetString("memory")
		cpuShare, _ := cmd.Flags().GetString("cpu-share")
		instances, _ := cmd.Flags().GetInt32("instances")
		signers, _ := cmd.Flags().GetInt32("signers")
		sigThreshold, _ := cmd.Flags().GetInt32("sig-threshold")

		ops := new(cocoon.Ops)
		err := ops.Create(&client.Cocoon{
			URL:          url,
			Language:     lang,
			ReleaseTag:   releaseTag,
			BuildParam:   buildParams,
			Memory:       memory,
			CPUShare:     cpuShare,
			Instances:    instances,
			Signers:      signers,
			SigThreshold: sigThreshold,
		})
		if err != nil {
			log.Fatalf("%s", common.StripRPCErrorPrefix([]byte(err.Error())))
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("url", "u", "", "Deploys a cocoon code to the platform")
	createCmd.Flags().StringP("lang", "l", "", "The langauges the cocoon code is written in")
	createCmd.Flags().StringP("release-tag", "r", "", "The github release tag. Defaults to `latest`")
	createCmd.Flags().StringP("build-param", "b", "", "Build parameters to apply during cocoon code build process")
	createCmd.Flags().StringP("memory", "m", "512m", "The amount of memory to allocate. e.g 512m, 1g or 2g")
	createCmd.Flags().StringP("cpu-share", "c", "1x", "The share of cpu to allocate. e.g 1x or 2x")
	createCmd.Flags().Int32P("instances", "i", 1, "The number of instances to run")
	createCmd.Flags().Int32P("signers", "s", 1, "The number of signatories")
	createCmd.Flags().Int32P("sig-threshold", "t", 1, "The number of signatures required confirm an update")
}
