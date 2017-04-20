package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/connector"
	"github.com/ncodes/cocoon/core/connector/server"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/spf13/cobra"
)

var (
	log = config.MakeLogger("connector", fmt.Sprintf("cocoon_%s", os.Getenv("COCOON_ID")))

	// connector RPC API
	defaultConnectorRPCAPI = util.Env("DEV_ADDR_CONNECTOR_RPC", ":8002")

	// Signals channel
	sigs = make(chan os.Signal, 1)
)

func init() {
	config.ConfigureLogger()
}

// creates a deployment request with argument
// fetched from the environment.
func getRequest() (*connector.Request, error) {

	// get cocoon code github link and language
	ccID := os.Getenv("COCOON_ID")
	ccURL := os.Getenv("COCOON_CODE_URL")
	ccVersion := os.Getenv("COCOON_CODE_VERSION")
	ccLang := os.Getenv("COCOON_CODE_LANG")
	diskLimit := util.Env("COCOON_DISK_LIMIT", "300")
	buildParam := os.Getenv("COCOON_BUILD_PARAMS")
	ccLink := os.Getenv("COCOON_LINK")
	memoryMB, _ := strconv.Atoi(util.Env("ALLOC_MEMORY", "4"))
	cpuShare, _ := strconv.Atoi(util.Env("ALLOC_CPU_SHARE", "100"))

	if ccID == "" {
		return nil, fmt.Errorf("Cocoon code id not set @ $COCOON_ID")
	} else if ccURL == "" {
		return nil, fmt.Errorf("Cocoon code url not set @ $COCOON_CODE_URL")
	} else if ccLang == "" {
		return nil, fmt.Errorf("Cocoon code url not set @ $COCOON_CODE_LANG")
	}

	return &connector.Request{
		ID:          ccID,
		URL:         ccURL,
		Version:     ccVersion,
		Lang:        ccLang,
		DiskLimit:   common.MBToByte(util.ToInt64(diskLimit)),
		BuildParams: buildParam,
		Link:        ccLink,
		Memory:      int64(memoryMB),
		CPUShare:   int64(cpuShare),
	}, nil
}

// onTerminate calls a function when a terminate or interrupt signal is received.
func onTerminate(f func(s os.Signal)) {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		f(s)
	}()
}

// startCmd represents the connector command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the connector and launch a cocoon code",
	Long:  `Starts the connector and launch a cocoon code`,
	Run: func(cmd *cobra.Command, args []string) {

		waitCh := make(chan bool, 1)
		serverStartedCh := make(chan bool)

		log.Info("Connector has started")

		// get request
		log.Info("Collecting and validating launch request")
		req, err := getRequest()
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Infof("Ready to launch cocoon code with id = %s", req.ID)

		connectorRPCAddr := scheduler.Getenv("ADDR_RPC", defaultConnectorRPCAPI)
		cocoonCodeRPCAddr := scheduler.Getenv("ADDR_code_RPC", ":8004")

		cn := connector.NewConnector(req, waitCh)
		cn.SetAddrs(connectorRPCAddr, cocoonCodeRPCAddr)
		cn.AddLanguage(connector.NewGo(req))
		onTerminate(func(s os.Signal) {
			log.Info("Terminate signal received. Stopping connector")
			cn.Stop(false)
		})

		// start grpc API server
		rpcServer := server.NewRPCServer(cn)
		go rpcServer.Start(connectorRPCAddr, serverStartedCh)

		// wait for rpc server to start then launch the cocoon code
		<-serverStartedCh
		go cn.Launch(connectorRPCAddr, cocoonCodeRPCAddr)

		if <-waitCh {
			rpcServer.Stop(1)
		} else {
			rpcServer.Stop(0)
		}

		log.Info("Connector stopped")
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
