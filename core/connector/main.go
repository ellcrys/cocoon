package main

import (
	"os"
	"time"

	"google.golang.org/grpc/grpclog"

	_ "net/http/pprof"

	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/cmd"
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
