package client

import (
	"fmt"

	"time"

	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("client")

// CCClient represents a cocoon code GRPC client
// that interacts with a cocoon code.
type CCClient struct {
	ccodePort int
}

// NewCCClient creates a new cocoon code client
func NewCCClient(ccodePort int) *CCClient {
	return &CCClient{
		ccodePort: ccodePort,
	}
}

// Connect connects to a cocoon code server
// running on a unknow port
func (c *CCClient) Connect() error {

	log.Info("Starting cocoon code client")

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", c.ccodePort), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to cocoon code server. %s", err)
	}
	defer conn.Close()

	log.Debugf("Now connected to cocoon code at port=%d", c.ccodePort)

	if err = c.Do(conn); err != nil {
		return err
	}

	return nil
}

// isServerAlive performs a health check
// against the server to ensure it is alive and running.
func (c *CCClient) isServerAlive() {

}

// Do starts a request loop that continously asks the
// server for transactions. When it receives a transaction,
// it processes it and returns the result.
func (c *CCClient) Do(conn *grpc.ClientConn) error {
	log.Info("Started transaction processing loop")

	for {
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
