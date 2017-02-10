package main

import (
	"os"

	"github.com/ncodes/cocoon/core/cmd"
	"github.com/ncodes/cocoon/core/config"
	"github.com/op/go-logging"
)

var log *logging.Logger

func init() {
	config.ConfigureLogger()
	log = logging.MustGetLogger("main")
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}
