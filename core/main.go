package main

import (
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/cmd"
	"github.com/ncodes/cocoon/core/config"
	"github.com/op/go-logging"
)

var log *logging.Logger

func init() {
	config.ConfigureLogger()
	log = logging.MustGetLogger("main")
	goreq.SetConnectTimeout(5 * time.Second)
}

func main() {
	// defer profile.Start(profile.MemProfile).Stop()

	// go func() {
	// 	go func() {
	// 		log.Info(http.ListenAndServe("localhost:6060", nil).Error())
	// 	}()

	// 	for {
	// 		var mem runtime.MemStats
	// 		runtime.ReadMemStats(&mem)
	// 		log.Infof("Alloc: %d, Total Alloc: %d, HeapAlloc: %d, HeapSys: %d", mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }()

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
}
