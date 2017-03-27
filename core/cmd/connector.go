package cmd

import (
	"fmt"
	"os"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector"
	"github.com/ncodes/cocoon/core/connector/server"
	"github.com/ncodes/cocoon/core/scheduler"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
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
	ccTag := os.Getenv("COCOON_CODE_TAG")
	ccLang := os.Getenv("COCOON_CODE_LANG")
	diskLimit := util.Env("COCOON_DISK_LIMIT", "300")
	buildParam := os.Getenv("COCOON_BUILD_PARAMS")
	ccLink := os.Getenv("COCOON_LINK")
	ccAddr := fmt.Sprintf(":%s", scheduler.Getenv("PORT_COCOON_RPC", "8000"))

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
		Tag:         ccTag,
		Lang:        ccLang,
		DiskLimit:   common.MBToByte(util.ToInt64(diskLimit)),
		BuildParams: buildParam,
		Link:        ccLink,
		CocoonAddr:  ccAddr,
	}, nil
}

// connectorCmd represents the connector command
var connectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Start the connector",
	Long:  `Starts the connector and launches a cocoon code.`,
	Run: func(cmd *cobra.Command, args []string) {

		waitCh := make(chan bool, 1)
		serverStartedCh := make(chan bool)

		var log = logging.MustGetLogger("connector")
		log.Info("Connector has started")

		// get request
		log.Info("Collecting and validating launch request")
		req, err := getRequest()
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Infof("Ready to launch cocoon code with id = %s", req.ID)

		cn := connector.NewConnector(waitCh)
		cn.AddLanguage(connector.NewGo(req))

		// start grpc API server
		rpcServer := server.NewRPCServer(cn)
		addr := scheduler.Getenv("ADDR_CONNECTOR_RPC", "127.0.0.1:8002")
		go rpcServer.Start(addr, serverStartedCh, make(chan bool, 1))
		<-serverStartedCh
		go cn.Launch(req, addr)

		if <-waitCh {
			rpcServer.Stop(1)
		} else {
			rpcServer.Stop(0)
		}

		log.Info("connector stopped")
	},
}

func init() {
	RootCmd.AddCommand(connectorCmd)
}
