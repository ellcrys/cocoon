package main

import (
	"os"
	"time"

	"google.golang.org/grpc/grpclog"

	_ "net/http/pprof"

	"github.com/franela/goreq"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	"github.com/ellcrys/cocoon/core/connector/cmd"
	"github.com/op/go-logging"
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
