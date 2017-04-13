package cmd

import (
	"time"

	"os"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [OPTIONS]",
	Short: "Create a new cocoon configuration locally",
	Long:  `Create a cocoon on the blockchain`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		url, _ := cmd.Flags().GetString("url")
		lang, _ := cmd.Flags().GetString("lang")
		releaseTag, _ := cmd.Flags().GetString("release-tag")
		buildParams, _ := cmd.Flags().GetString("build-param")
		memory, _ := cmd.Flags().GetString("memory")
		cpuShare, _ := cmd.Flags().GetString("cpu-share")
		link, _ := cmd.Flags().GetString("link")
		numSig, _ := cmd.Flags().GetInt32("num-sig")
		sigThreshold, _ := cmd.Flags().GetInt32("sig-threshold")
		firewall, _ := cmd.Flags().GetString("firewall")

		// validate firewall
		if len(firewall) > 0 {
			firewallValidationErrs := client.ValidateFirewall(firewall)
			if firewallValidationErrs != nil && len(firewallValidationErrs) > 0 {
				for _, err := range firewallValidationErrs {
					log.Infof("Err: firewall: %s", err.Error())
				}
				os.Exit(1)
			}
		}

		err := client.CreateCocoon(&types.Cocoon{
			ID:             util.UUID4(),
			URL:            url,
			Language:       lang,
			ReleaseTag:     releaseTag,
			BuildParam:     buildParams,
			Memory:         memory,
			CPUShares:      cpuShare,
			Link:           link,
			NumSignatories: numSig,
			SigThreshold:   sigThreshold,
			CreatedAt:      time.Now().UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("url", "u", "", "The github repository url of the cocoon code")
	createCmd.Flags().StringP("lang", "l", "", "The langauges the cocoon code is written in")
	createCmd.Flags().StringP("release-tag", "r", "", "The github release tag. Defaults to `latest`")
	createCmd.Flags().StringP("firewall", "f", "", "The outgoing firewall rules of the cocoon")
	createCmd.Flags().StringP("build-param", "b", "", "Build parameters to apply during cocoon code build process")
	createCmd.Flags().StringP("memory", "m", "512m", "The amount of memory to allocate. e.g 512m, 1g or 2g")
	createCmd.Flags().StringP("link", "", "", "The id of an existing cocoon to natively link to.")
	createCmd.Flags().StringP("cpu-share", "c", "1x", "The share of cpu to allocate. e.g 1x or 2x")
	createCmd.Flags().Int32P("num-sig", "s", 1, "The number of signatories")
	createCmd.Flags().Int32P("sig-threshold", "t", 1, "The number of signatures required to confirm a new release")
}
