package client

import (
	"fmt"

	"io"

	"os"

	"time"

	proto "github.com/ncodes/cocoon/core/stubs/golang/proto"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("connector.client")

// Client represents a cocoon code GRPC client
// that interacts with a cocoon code.
type Client struct {
	ccodePort        string
	stub             proto.StubClient
	conCtx           context.Context
	conCancel        context.CancelFunc
	orderDiscoTicker *time.Ticker
	orderersAddr     []string
}

// NewClient creates a new cocoon code client
func NewClient(ccodePort string) *Client {
	return &Client{
		ccodePort: ccodePort,
	}
}

// GetCCPort returns the cocoon code port.
// For development, if DEV_COCOON_CODE_PORT is set, it connects to it.
func (c *Client) GetCCPort() string {
	if devCCodePort := os.Getenv("DEV_COCOON_CODE_PORT"); len(devCCodePort) > 0 {
		return devCCodePort
	}
	return c.ccodePort
}

// Connect connects to a cocoon code server
// running on a known port
func (c *Client) Connect() error {

	log.Info("Starting cocoon code client")

	// start a ticker to continously discover orderer addreses
	go func() {
		c.orderDiscoTicker = time.NewTicker(60 * time.Second)
		for _ = range c.orderDiscoTicker.C {
			c.discoverOrderers()
		}
	}()

	c.discoverOrderers()
	if len(c.orderersAddr) > 0 {
		log.Infof("Orderer address list updated. Contains %d orderer address(es)", len(c.orderersAddr))
	} else {
		log.Warning("No orderer address was found. We won't be able to reach the orderer. ")
	}

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%s", c.GetCCPort()), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to cocoon code server. %s", err)
	}
	defer conn.Close()

	log.Debugf("Now connected to cocoon code at port=%s", c.GetCCPort())

	c.stub = proto.NewStubClient(conn)

	if err = c.Do(conn); err != nil {
		return err
	}

	return nil
}

// discoverOrderers fetches a list of orderer service addresses
// via consul service discovery API. For development purpose,
// If DEV_ORDERER_ADDR is set, it will fetch the orderer
// address from the env variable.
func (c *Client) discoverOrderers() {
	if len(os.Getenv("DEV_ORDERER_ADDR")) > 0 {
		c.orderersAddr = []string{os.Getenv("DEV_ORDERER_ADDR")}
	}
	// Retrieve from consul service API (not implemented)
}

// Do starts a request loop that continously asks the
// server for transactions. When it receives a transaction,
// it processes it and returns the result.
func (c *Client) Do(conn *grpc.ClientConn) error {

	// create a context so we have complete controll of the connection
	c.conCtx, c.conCancel = context.WithCancel(context.Background())

	// connect to the cocoon code
	stream, err := c.stub.Transact(c.conCtx)
	if err != nil {
		return fmt.Errorf("failed to call GetTx of cocoon code stub. %s", err)
	}

	for {

		log.Info("Waiting for transactions")

		in, err := stream.Recv()
		if err == io.EOF {
			return fmt.Errorf("GetTx connection between connector and cocoon code stub has ended")
		}
		if err != nil {
			return fmt.Errorf("Failed to successfully receive message from cocoon code. %s", err)
		}

		log.Infof("New Tx From Cocoon: %", in.String())
	}
}
