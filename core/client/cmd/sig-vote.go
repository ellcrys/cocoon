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
	"github.com/ncodes/cocoon/core/client/client"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// sigVoteCmd represents the sig-vote command
var sigVoteCmd = &cobra.Command{
	Use:   "sig-vote [OPTIONS] RELEASE VOTE",
	Short: "Vote to approve or deny a cocoon release",
	Long:  `Vote to approve or deny a cocoon release`,
	Run: func(cmd *cobra.Command, args []string) {

		isCid, _ := cmd.Flags().GetBool("cid")
		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(args) <= 1 {
			UsageError(log, cmd, `"ellcrys sig-vote" requires at least 2 argument(s)`, `ellcrys sig-vote --help`)
		} else if args[1] != "0" && args[1] != "1" {
			log.Fatal("Err: Vote value must be either 1 (Approve) or 0 (Deny)")
		}

		if err := client.AddVote(args[0], args[1], isCid); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(sigVoteCmd)
	sigVoteCmd.PersistentFlags().BoolP("cid", "", false, "Force release ID to be interpreted as a cocoon ID (Default: false)")
}
