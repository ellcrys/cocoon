package client

import (
	"fmt"

	"io"

	proto "github.com/ncodes/cocoon/core/stubs/golang/proto"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("connector.client")

// Client represents a cocoon code GRPC client
// that interacts with a cocoon code.
type Client struct {
	ccodePort int
	stub      proto.StubClient
	conCtx    context.Context
	conCancel context.CancelFunc
}

// NewClient creates a new cocoon code client
func NewClient(ccodePort int) *Client {
	return &Client{
		ccodePort: ccodePort,
	}
}

// Connect connects to a cocoon code server
// running on a known port
func (c *Client) Connect() error {

	log.Info("Starting cocoon code client")

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", c.ccodePort), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to cocoon code server. %s", err)
	}
	defer conn.Close()

	log.Debugf("Now connected to cocoon code at port=%d", c.ccodePort)

	c.stub = proto.NewStubClient(conn)

	if err = c.Do(conn); err != nil {
		return err
	}

	return nil
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
