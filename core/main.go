package main

import (
	"net/http"
	"os"
	"runtime"
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

	go func() {
		go func() {
			log.Info(http.ListenAndServe("localhost:6060", nil))
		}()

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		log.Infof("Alloc: %s, Total Alloc: %s, HeapAlloc: %s, HeapSys: %s", mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
	}()

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
}
