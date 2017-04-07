package cmd

import (
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [OPTIONS] COCOON",
	Short: "Update a cocoon and/or create a new release",
	Long:  `Update a cocoon and/or create a new release`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys update" requires at least 1 argument(s)`, `ellcrys update --help`)
		}

		url, _ := cmd.Flags().GetString("url")
		lang, _ := cmd.Flags().GetString("lang")
		releaseTag, _ := cmd.Flags().GetString("release-tag")
		buildParams, _ := cmd.Flags().GetString("build-param")
		memory, _ := cmd.Flags().GetString("memory")
		cpuShare, _ := cmd.Flags().GetString("cpu-share")
		link, _ := cmd.Flags().GetString("link")
		numSig, _ := cmd.Flags().GetInt32("num-sig")
		sigThreshold, _ := cmd.Flags().GetInt32("sig-threshold")

		upd := &proto.CocoonPayloadRequest{
			ID:             args[0],
			URL:            url,
			Language:       lang,
			ReleaseTag:     releaseTag,
			BuildParam:     buildParams,
			Memory:         memory,
			CPUShares:      cpuShare,
			Link:           link,
			NumSignatories: numSig,
			SigThreshold:   sigThreshold,
		}

		if err := client.UpdateCocoon(args[0], upd); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("url", "u", "", "The github repository url of the cocoon code")
	updateCmd.Flags().StringP("lang", "l", "", "The langauges the cocoon code is written in")
	updateCmd.Flags().StringP("release-tag", "r", "", "The github release tag. Defaults to `latest`")
	updateCmd.Flags().StringP("build-param", "b", "", "Build parameters to apply during cocoon code build process")
	updateCmd.Flags().StringP("memory", "m", "", "The amount of memory to allocate. e.g 512m, 1g or 2g")
	updateCmd.Flags().StringP("link", "", "", "The id of an existing cocoon to natively link to.")
	updateCmd.Flags().StringP("cpu-share", "c", "", "The share of cpu to allocate. e.g 1x or 2x")
	updateCmd.Flags().Int32P("num-sig", "s", 0, "The number of signatories")
	updateCmd.Flags().Int32P("sig-threshold", "t", 0, "The number of signatures required to confirm a new release")
}
