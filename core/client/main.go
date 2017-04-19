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

package main

import (
	"os"
	"path"

	"google.golang.org/grpc/grpclog"

	"github.com/mitchellh/go-homedir"
	"github.com/ncodes/cocoon/core/client/cmd"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
)

var log *logging.Logger

// ProjectName is the official name of the project
var ProjectName = "cocoon"

// createConfigDir creates the config directory if it doesn't exists
func createConfigDir() {

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("failed to determine home directory")
	}

	projectConfigDir := path.Join(home, ".config", ProjectName)
	if _, err = os.Stat(projectConfigDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(projectConfigDir, 0777); err != nil {
				log.Fatal("failed to create config directory")
			}
		}
	}
}

func init() {
	config.ConfigureLogger()
	log = logging.MustGetLogger("api.client")
	createConfigDir()
	log.SetBackend(config.MessageOnlyBackend)
	if len(os.Getenv("ENABLE_GRPC_LOG")) == 0 {
		gl := common.GLogger{}
		gl.Disable(true, true)
		grpclog.SetLogger(&gl)
	}
}

func main() {
	cmd.Execute()
}
