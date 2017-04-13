package client

import (
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/config"
)

// APIAddress is the remote address to the cluster server
var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8004")

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}
