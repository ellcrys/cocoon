package main

import (
	"os"
	"time"

	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/cmd"
	"github.com/ncodes/cocoon/core/config"
	"github.com/op/go-logging"
	"github.com/pkg/profile"
)

var log *logging.Logger

func init() {
	config.ConfigureLogger()
	log = logging.MustGetLogger("main")
	goreq.SetConnectTimeout(5 * time.Second)
}

func main() {
	defer profile.Start().Stop()
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
}
