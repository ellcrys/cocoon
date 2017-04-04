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

// sigAddCmd represents the sig-add command
var sigAddCmd = &cobra.Command{
	Use:   "sig-add",
	Short: "Add a signatory to a cocoon",
	Long:  `Add a signatory to a cocoon`,
	Run: func(cmd *cobra.Command, args []string) {

		id, _ := cmd.Flags().GetString("id")
		log := logging.MustGetLogger("api.client")
		log.SetBackend(config.MessageOnlyBackend)

		if len(id) == 0 {
			log.Fatal("Err: Cocoon ID is required")
		}
		if len(args) == 0 {
			log.Fatal("Err: One or more identity IDs are required")
		}

		if err := client.AddSignatories(id, args); err != nil {
			log.Fatalf("Err: %s", common.CapitalizeString((common.GetRPCErrDesc(err))))
		}
	},
}

func init() {
	RootCmd.AddCommand(sigAddCmd)
	sigAddCmd.PersistentFlags().String("id", "", "The cocoon id")
}
