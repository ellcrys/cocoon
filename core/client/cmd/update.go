package cmd

import (
	"os"

	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cstructs"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [OPTIONS] CONTRACT_FILE_PATH",
	Short: "Update a cocoon and/or create a new release",
	Long:  `Update a cocoon and/or create a new release`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys update" requires at least 1 argument(s)`, `ellcrys update --help`)
		}

		cocoons, errs := parseContract(args[0])
		if errs != nil && len(errs) > 0 {
			for _, err := range errs {
				log.Errorf("Err: %s", common.CapitalizeString(err.Error()))
			}
			os.Exit(1)
		}

		for i, cocoon := range cocoons {
			var protoCreatePayloadReq proto.CocoonPayloadRequest
			cstructs.Copy(cocoon, &protoCreatePayloadReq)
			protoCreatePayloadReq.ACL = cocoon.ACL.ToJSON()
			err := client.UpdateCocoon(protoCreatePayloadReq.ID, &protoCreatePayloadReq)
			if err != nil {
				log.Fatalf("Err (Contract %d): %s", i, common.CapitalizeString((common.GetRPCErrDesc(err))))
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
