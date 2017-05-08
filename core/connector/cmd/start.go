package cmd

import (
	"os"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/connector"
	"github.com/ncodes/cocoon/core/connector/connector/languages"
	"github.com/ncodes/cocoon/core/connector/router"
	"github.com/ncodes/cocoon/core/connector/server"
	"github.com/ncodes/cocoon/core/platform"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	log       = config.MakeLogger("connector")
	routerLog = config.MakeLogger("connector.router_helper")

	// default connector RPC Addr
	defaultConnectorRPCAddr = util.Env("DEV_ADDR_CONNECTOR_RPC", ":8002")

	// default connector HTTP Addr
	defaultConnectorHTTPAddr = util.Env("DEV_ADDR_CONNECTOR_HTTP", ":8900")

	// default cocoon code RPC ADDR
	defaultCocoonCodeRPCAddr = util.Env("DEV_ADDR_COCOON_CODE_RPC", "127.0.0.1:8004")
)

func init() {
	config.ConfigureLogger()
}

// Pull the cocoon specification
func getSpec(pf *platform.Platform) (*types.Spec, error) {

	ctx, cc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cc()

	cocoon, release, err := pf.GetCocoonAndRelease(ctx, os.Getenv("COCOON_ID"), os.Getenv("COCOON_RELEASE"), true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get spec")
	}

	diskLimit := util.Env("COCOON_DISK_LIMIT", "300")
	return &types.Spec{
		ID:          cocoon.ID,
		URL:         release.URL,
		Version:     release.Version,
		Lang:        release.Language,
		DiskLimit:   common.MBToByte(util.ToInt64(diskLimit)),
		BuildParams: release.BuildParam,
		Link:        release.Link,
		ReleaseID:   release.ID,
		Cocoon:      cocoon,
		Release:     release,
	}, nil
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
			"COCOON_RELEASE",
			"COCOON_CONTAINER_NAME",
		}...); len(missingEnv) > 0 {
			log.Fatalf("The following environment variables must be set: %v", missingEnv)
		}

		platform, err := platform.NewPlatform()
		if err != nil {
			log.Fatalf("Failed to initialize platform: %s", err)
			return
		}

		log.Infof("Connector initialized [cocoon=%s, release=%s]", os.Getenv("COCOON_ID"), os.Getenv("COCOON_RELEASE"))
		log.Infof("Fetching cocoon specification")
		spec, err := getSpec(platform)
		if err != nil {
			log.Fatalf("%s", err)
		}

		log.Infof("Launching...")

		connectorRPCAddr := scheduler.Getenv("ADDR_RPC", defaultConnectorRPCAddr)
		connectorHTTPAddr := scheduler.Getenv("ADDR_HTTP", defaultConnectorHTTPAddr)
		cocoonCodeRPCAddr := scheduler.Getenv("ADDR_code_RPC", defaultCocoonCodeRPCAddr)

		// create router helper
		routerHelper, err := router.NewHelper(routerLog, connectorHTTPAddr)
		if err != nil {
			log.Error(err.Error())
			log.Fatal("Ensure consul is running at 127.0.0.1:8500. Use CONSUL_ADDR to set alternative consul address")
		}

		// initialize connector
		waitCh := make(chan bool, 1)
		cn := connector.NewConnector(platform, spec, waitCh)
		cn.SetRouterHelper(routerHelper)
		cn.SetAddrs(connectorRPCAddr, cocoonCodeRPCAddr)
		cn.AddLanguage(languages.NewGo(spec))

		// start grpc API server
		rpcServerStartedCh := make(chan bool)
		rpcServer := server.NewRPC(cn)

		// start http server
		httpServerStartedCh := make(chan bool)
		httpServer := server.NewHTTP(rpcServer)

		go rpcServer.Start(connectorRPCAddr, rpcServerStartedCh)
		<-rpcServerStartedCh

		go httpServer.Start(connectorHTTPAddr, httpServerStartedCh)
		<-httpServerStartedCh

		// listen to terminate request
		common.OnTerminate(func(s os.Signal) {
			log.Info("Terminate signal received")
			log.Info("Stopping cocoon code")
			rpcServer.Stop()
			// allow some time for logs to be read by the connector
			time.Sleep(2 * time.Second)
			log.Info("Stopping connector")
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
