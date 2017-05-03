package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/connector"
	"github.com/ncodes/cocoon/core/connector/router"
	"github.com/ncodes/cocoon/core/connector/server"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/spf13/cobra"
)

var (
	log       = config.MakeLogger("connector", fmt.Sprintf("cocoon_%s", os.Getenv("COCOON_ID")))
	routerLog = config.MakeLogger("connector.routerHelper", "routerHelper")

	// default connector RPC Addr
	defaultConnectorRPCAddr = util.Env("DEV_ADDR_CONNECTOR_RPC", ":8002")

	// default connector HTTP Addr
	defaultConnectorHTTPAddr = util.Env("DEV_ADDR_CONNECTOR_HTTP", ":8900")

	// default cocoon code RPC ADDR
	defaultCocoonCodeRPCAddr = util.Env("DEV_ADDR_COCOON_CODE_RPC", ":8004")

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

		if missingEnv := common.HasEnv([]string{
			"ROUTER_DOMAIN",
			"COCOON_ID",
			"COCOON_CODE_URL",
			"COCOON_CODE_LANG",
		}...); len(missingEnv) > 0 {
			log.Fatalf("The following environment variables must be set: %v", missingEnv)
		}

		log.Info("Connector has started")

		// get deployment request from environment and validate it
		log.Info("Collecting and validating launch request")
		req, err := getRequest()
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Infof("Ready to launch cocoon code with id = %s", req.ID)

		connectorRPCAddr := scheduler.Getenv("ADDR_RPC", defaultConnectorRPCAddr)
		connectorHTTPAddr := scheduler.Getenv("ADDR_HTTP", defaultConnectorHTTPAddr)
		cocoonCodeRPCAddr := scheduler.Getenv("ADDR_code_RPC", defaultCocoonCodeRPCAddr)

		// create router helper
		routerHelper, err := router.NewHelper(routerLog, connectorHTTPAddr)
		if err != nil {
			log.Error(err.Error())
			log.Fatal("Ensure consul is running at 127.0.0.1:8500. Use CONSUL_ADDR to set alternative consul address")
		}

		waitCh := make(chan bool, 1)
		cn, err := connector.NewConnector(req, waitCh)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		cn.SetRouterHelper(routerHelper)
		cn.SetAddrs(connectorRPCAddr, cocoonCodeRPCAddr)
		cn.AddLanguage(connector.NewGo(req))

		// start grpc API server
		rpcServerStartedCh := make(chan bool)
		rpcServer := server.NewRPC(cn)
		go rpcServer.Start(connectorRPCAddr, rpcServerStartedCh)
		<-rpcServerStartedCh

		// start http server
		httpServerStartedCh := make(chan bool)
		httpServer := server.NewHTTP(rpcServer)
		go httpServer.Start(connectorHTTPAddr, httpServerStartedCh)
		<-httpServerStartedCh

		// listen to terminate request
		onTerminate(func(s os.Signal) {
			log.Info("Terminate signal received. Stopping connector")
			rpcServer.Stop()
			cn.Stop(false)
		})

		// launch the deployed cocoon code
		go cn.Launch(connectorRPCAddr, cocoonCodeRPCAddr)
		<-waitCh
		log.Info("Connector stopped")
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
