package cmd

import (
	"github.com/ncodes/cocoon/core/cluster"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

var log = logging.MustGetLogger("deploy")

// Deploy calls the clusters Deploy method to
// start a new cocoon.
func Deploy(cluster cluster.Cluster, jobID, lang, url, tag, buildParams string) (string, error) {
	return cluster.Deploy(jobID, lang, url, tag, buildParams)
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a smart contract to cocoon cluster",
	Long:  `This command deploys a smart contract to the cocoon cluster`,
	Run: func(cmd *cobra.Command, args []string) {

		id, _ := cmd.Flags().GetString("id")
		lang, _ := cmd.Flags().GetString("lang")
		url, _ := cmd.Flags().GetString("url")
		tag, _ := cmd.Flags().GetString("tag")
		buildParams, _ := cmd.Flags().GetString("build-params")
		clusterAddr, _ := cmd.Flags().GetString("cluster-addr")
		clusterAddrHTTPS, _ := cmd.Flags().GetBool("cluster-addr-https")

		cl := cluster.NewNomad()
		cl.SetAddr(clusterAddr, clusterAddrHTTPS)
		cocoonID, err := Deploy(cl, id, lang, url, tag, buildParams)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Info("Deployment ID:", cocoonID)
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("id", "", "", "The id of the job")
	deployCmd.Flags().StringP("lang", "l", "go", "The smart contract language")
	deployCmd.Flags().StringP("url", "u", "", "A zip file or github link to the smart contract")
	deployCmd.Flags().StringP("tag", "t", "", "The github release tag")
	deployCmd.Flags().StringP("cluster-addr", "", "127.0.0.1:4646", "The cluster address as host:port")
	deployCmd.Flags().BoolP("cluster-addr-https", "", false, "Whether to include `https` when accessing cluster APIs")
	deployCmd.Flags().StringP("build-params", "", "", "Specify build parameters")

	deployCmd.MarkFlagRequired("lang")
	deployCmd.MarkFlagRequired("url")
}
