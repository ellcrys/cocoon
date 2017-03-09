package cmd

import (
	"github.com/ncodes/cocoon/core/client/cocoon"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/types/client"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cocoon",
	Long:  `Create a cocoon`,
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

		ops := new(cocoon.Ops)
		err := ops.Create(&client.Cocoon{
			URL:        url,
			Lang:       lang,
			ReleaseTag: releaseTag,
			BuildParam: buildParams,
			Memory:     memory,
			CPUShare:   cpuShare,
			Instances:  instances,
		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("url", "u", "", "Deploys a cocoon code to the platform")
	createCmd.Flags().StringP("lang", "l", "", "The langauges the cocoon code is written in")
	createCmd.Flags().StringP("release-tag", "", "", "The github release tag. Defaults to `latest`")
	createCmd.Flags().StringP("build-param", "", "", "Build parameters to apply during cocoon code build process")
	createCmd.Flags().StringP("memory", "m", "512m", "The amount of memory to allocate. e.g 512m, 1g or 2g")
	createCmd.Flags().StringP("cpu-share", "c", "1x", "The share of cpu to allocate. e.g 1x or 2x")
	createCmd.Flags().Int32P("instances", "i", 1, "The number of instances to run")
}
