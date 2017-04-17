package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [OPTIONS] COCOON",
	Short: "Display recent log output from a cocoon",
	Long:  `Display recent log output from a cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys logs" requires at least 1 argument(s)`, `ellcrys logs --help`)
		}

		numLines, _ := cmd.Flags().GetInt("num")
		stderrOnly, _ := cmd.Flags().GetBool("stderr")
		stdoutOnly, _ := cmd.Flags().GetBool("stdout")
		tail, _ := cmd.Flags().GetBool("tail")

		if err := client.GetLogs(args[0], numLines, tail, stderrOnly, stdoutOnly); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)
	logsCmd.PersistentFlags().IntP("num", "n", client.MinimumLogLines, "Number of lines to display")
	logsCmd.PersistentFlags().Bool("stderr", false, "Display only stderr logs")
	logsCmd.PersistentFlags().Bool("stdout", false, "Display only stdout logs")
	logsCmd.PersistentFlags().BoolP("tail", "t", false, "Continually stream logs")
}
