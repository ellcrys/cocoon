// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"

	"github.com/ncodes/cocoon/core/cluster"
	"github.com/ncodes/cocoon/core/pod"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a smart contract to cocoon cluster",
	Long:  `This command deploys a smart contract to the cocoon cluster`,
	Run: func(cmd *cobra.Command, args []string) {

		lang, _ := cmd.Flags().GetString("lang")
		url, _ := cmd.Flags().GetString("url")
		tag, _ := cmd.Flags().GetString("tag")
		clusterAddr, _ := cmd.Flags().GetString("cluster_addr")
		clusterAddrHTTPS, _ := cmd.Flags().GetBool("cluster_addr_https")

		cl := cluster.NewNomad()
		cl.SetAddr(clusterAddr, clusterAddrHTTPS)
		cocoonID, err := pod.Deploy(cl, lang, url, tag)
		log.Println(cocoonID, err)
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("lang", "l", "go", "The smart contract language")
	deployCmd.Flags().StringP("url", "u", "", "A zip file or github link to the smart contract")
	deployCmd.Flags().StringP("tag", "t", "", "The github release tag")
	deployCmd.Flags().StringP("cluster_addr", "", "127.0.0.1:4646", "The cluster address as host:port")
	deployCmd.Flags().BoolP("cluster_addr_https", "", false, "Whether to include `https` when accessing cluster APIs")

	deployCmd.MarkFlagRequired("lang")
	deployCmd.MarkFlagRequired("url")
}
