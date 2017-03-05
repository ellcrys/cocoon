package main

import (
	"os"
	"time"

	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/api/cmd"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
)

var log *logging.Logger

func init() {
	config.ConfigureLogger()
	log = logging.MustGetLogger("api")
	goreq.SetConnectTimeout(5 * time.Second)
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
}
