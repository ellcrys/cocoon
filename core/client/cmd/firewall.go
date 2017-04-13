package cmd

import (
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// firewallCmd represents the firewall command
var firewallCmd = &cobra.Command{
	Use:   "firewall [COMMAND]",
	Short: "Manage a cocoon's outgoing network traffic",
	Long:  `Manage a cocoon's outgoing network traffic`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// TODO: Work your own magic here
	// 	fmt.Println("firewall called")
	// },
}

var allowRuleCmd = &cobra.Command{
	Use:   "allow COCOON [OPTIONS]",
	Short: "Allow an outgoing connection",
	Long:  `Allow an outgoing connection`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		dest, _ := cmd.Flags().GetString("destination")
		protocol, _ := cmd.Flags().GetString("protocol")
		destPort, _ := cmd.Flags().GetString("port")

		if len(args) == 0 {
			UsageError(log, cmd, `"ellcrys firewall allow" requires at least 1 argument(s)`, `ellcrys firewall allow --help`)
		}
		if len(dest) == 0 {
			UsageError2(log, cmd, `Destination address is required`, `ellcrys firewall allow --help`)
		}
		if len(destPort) == 0 {
			UsageError2(log, cmd, `Destination port is required`, `ellcrys firewall allow --help`)
		}

		if err := client.FirewallAllow(dest, destPort, protocol); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	firewallCmd.AddCommand(allowRuleCmd)
	RootCmd.AddCommand(firewallCmd)

	allowRuleCmd.PersistentFlags().StringP("destination", "d", "", "The destination address or IP")
	allowRuleCmd.PersistentFlags().StringP("port", "p", "", "The destination port")
	allowRuleCmd.PersistentFlags().String("protocol", "tcp", "The connection protocol")
}
