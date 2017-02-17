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
	"fmt"
	"net"

	"time"

	"github.com/ncodes/cocoon/core/connector/ccode"
	"github.com/ncodes/cocoon/core/connector/config"
	"github.com/ncodes/cocoon/core/connector/proto"
	"github.com/ncodes/cocoon/core/connector/server"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	config.ConfigureLogger()
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the connector",
	Long:  `Starts the connector. Pulls chaincode specified in COCOON_CODE_URL, builds it and starts it.`,
	Run: func(cmd *cobra.Command, args []string) {

		doneCh := make(chan bool, 1)
		ccInstallFailedCh := make(chan bool, 1)

		port, _ := cmd.Flags().GetString("port")
		var log = logging.MustGetLogger("connector")
		var opts []grpc.ServerOption
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			log.Errorf("Error creating network interface to listen on. %s", err)
		}

		log.Info("Starting GRPC server")
		time.AfterFunc(2*time.Second, func() {
			log.Infof("GRPC server started on port %s", port)

			// install cooncode
			ccode.Install(ccInstallFailedCh)
		})

		// time.AfterFunc()
		grpcServer := grpc.NewServer(opts...)
		proto.RegisterConnectorServer(grpcServer, server.NewServer())
		go grpcServer.Serve(lis)

		httpServer := server.NewHTTPServer()
		go httpServer.Start("3000")

		if <-ccInstallFailedCh {
			log.Fatal("aborting: chaincode installation failed")
		}

		// start cocoon code
		ccode.Start()

		<-doneCh
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("port", "p", "5500", "The port to run bind to")
}
