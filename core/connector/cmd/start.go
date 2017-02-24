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
	"os"

	"fmt"

	"github.com/ncodes/cocoon/core/connector/config"
	"github.com/ncodes/cocoon/core/connector/launcher"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

func init() {
	config.ConfigureLogger()
}

// creates a deployment request with argument
// fetched from the environment.
func getRequest() (*launcher.Request, error) {

	// get cocoon code github link and language
	ccID := os.Getenv("COCOON_ID")
	ccURL := os.Getenv("COCOON_CODE_URL")
	ccTag := os.Getenv("COCOON_CODE_TAG")
	ccLang := os.Getenv("COCOON_CODE_LANG")

	if ccID == "" {
		return nil, fmt.Errorf("Cocoon code id not set @ $COCOON_ID")
	} else if ccURL == "" {
		return nil, fmt.Errorf("Cocoon code url not set @ $COCOON_CODE_URL")
	} else if ccLang == "" {
		return nil, fmt.Errorf("Cocoon code url not set @ $COCOON_CODE_LANG")
	}

	return &launcher.Request{
		ID:   ccID,
		URL:  ccURL,
		Tag:  ccTag,
		Lang: ccLang,
	}, nil
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the launcher",
	Long:  `Starts the launcher. Pulls cocoon code specified in COCOON_CODE_URL, builds it and runs it.`,
	Run: func(cmd *cobra.Command, args []string) {

		doneCh := make(chan bool, 1)
		launchFailedCh := make(chan bool, 1)

		var log = logging.MustGetLogger("connector")
		log.Info("Connector started. Initiating cocoon code launch procedure.")

		// get request
		req, err := getRequest()
		if err != nil {
			log.Error(err.Error())
			return
		}

		// install cooncode
		lchr := launcher.NewLauncher(launchFailedCh)
		lchr.AddLanguage(launcher.NewGo())
		lchr.Launch(req)

		if <-launchFailedCh {
			log.Fatal("aborting: cocoon code launch failed")
			lchr.Stop()
		}

		<-doneCh
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	// startCmd.Flags().StringP("port", "p", "5500", "The port to run bind to")
}
