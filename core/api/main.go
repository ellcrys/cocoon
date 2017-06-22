package main

import (
	"os"
	"time"

	"google.golang.org/grpc/grpclog"

	"github.com/ellcrys/cocoon/core/api/cmd"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	"github.com/franela/goreq"
	logging "github.com/op/go-logging"
)

var log *logging.Logger

func init() {
	config.ConfigureLogger()
	log = logging.MustGetLogger("main")
	goreq.SetConnectTimeout(5 * time.Second)
	if len(os.Getenv("ENABLE_GRPC_LOG")) == 0 {
		gl := common.GLogger{}
		gl.Disable(true, true)
		grpclog.SetLogger(&gl)
	}
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
}
