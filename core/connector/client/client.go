package client

import (
	"fmt"
	"io"
	"os"
	"time"

	"strings"

	"github.com/ncodes/cocoon/core/orderer"
	stub "github.com/ncodes/cocoon/core/stubs/golang"
	proto "github.com/ncodes/cocoon/core/stubs/golang/proto"
	logging "github.com/op/go-logging"
	cmap "github.com/orcaman/concurrent-map"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("connector.client")

// txChannels holds the channels to send transaction responses to
var txRespChannels = cmap.New()

// Client represents a cocoon code GRPC client
// that interacts with a cocoon code.
type Client struct {
	ccodePort        string
	stub             proto.StubClient
	conCtx           context.Context
	conCancel        context.CancelFunc
	orderDiscoTicker *time.Ticker
	orderersAddr     []string
	stream           proto.Stub_TransactClient
	cocoonID         string
}

// NewClient creates a new cocoon code client
func NewClient(ccodePort string) *Client {
	return &Client{
		ccodePort: ccodePort,
	}
}

// SetCocoonID sets the cocoon id
func (c *Client) SetCocoonID(id string) {
	c.cocoonID = id
}

// getCCPort returns the cocoon code port.
// For development, if DEV_COCOON_CODE_PORT is set, it connects to it.
func (c *Client) getCCPort() string {
	if devCCodePort := os.Getenv("DEV_COCOON_CODE_PORT"); len(devCCodePort) > 0 {
		return devCCodePort
	}
	return c.ccodePort
}

// Close the stream and cancel connections
func (c *Client) Close() {
	c.stream.CloseSend()
	c.conCancel()
}

// GetStream returns the grpc stream connected to the grpc cocoon code server
func (c *Client) GetStream() proto.Stub_TransactClient {
	return c.stream
}

// Connect connects to a cocoon code server
// running on a known port
func (c *Client) Connect() error {

	log.Info("Starting cocoon code client")

	// start a ticker to continously discover orderer addreses
	go func() {
		c.orderDiscoTicker = time.NewTicker(60 * time.Second)
		for _ = range c.orderDiscoTicker.C {
			c.orderersAddr = orderer.DiscoverOrderers()
		}
	}()

	c.orderersAddr = orderer.DiscoverOrderers()
	if len(c.orderersAddr) > 0 {
		log.Infof("Orderer address list updated. Contains %d orderer address(es)", len(c.orderersAddr))
	} else {
		log.Warning("No orderer address was found. We won't be able to reach the orderer. ")
	}

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%s", c.getCCPort()), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to cocoon code server. %s", err)
	}
	defer conn.Close()

	log.Debugf("Now connected to cocoon code at port=%s", c.getCCPort())

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

	var err error

	// create a context so we have complete controll of the connection
	c.conCtx, c.conCancel = context.WithCancel(context.Background())

	// connect to the cocoon code
	c.stream, err = c.stub.Transact(c.conCtx)
	if err != nil {
		return fmt.Errorf("failed to start transaction stream with cocoon code. %s", err)
	}

	for {

		in, err := c.stream.Recv()
		if err == io.EOF {
			return fmt.Errorf("connection with cocoon code has ended")
		}
		if err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				log.Info("Connection to cocoon code closed")
				return nil
			}
			return fmt.Errorf("failed to read message from cocoon code. %s", err)
		}

		switch in.Invoke {
		case true:
			go func() {
				log.Debugf("New invoke transaction (%s) from cocoon code", in.GetId())
				if err = c.handleInvokeTransaction(in); err != nil {
					log.Error(err.Error())
					c.stream.Send(&proto.Tx{
						Response: true,
						Id:       in.GetId(),
						Status:   500,
						Body:     []byte(err.Error()),
					})
				}
			}()
		case false:
			log.Debugf("New response transaction (%s) from cocoon code", in.GetId())
			go func() {
				if err = c.handleRespTransaction(in); err != nil {
					log.Error(err.Error())
					c.stream.Send(&proto.Tx{
						Response: true,
						Id:       in.GetId(),
						Status:   500,
						Body:     []byte(err.Error()),
					})
				}
			}()
		}
	}
}

// handleInvokeTransaction processes invoke transaction requests
func (c *Client) handleInvokeTransaction(tx *proto.Tx) error {
	switch tx.GetName() {
	case stub.TxCreateLedger:
		return c.createLedger(tx)
	case stub.TxPut:
		return c.put(tx)
	case stub.TxGetLedger:
		return c.getLedger(tx)
	case stub.TxGet:
		return c.get(tx)
	default:
		return c.stream.Send(&proto.Tx{
			Id:       tx.GetId(),
			Response: true,
			Status:   500,
			Body:     []byte(fmt.Sprintf("unsupported transaction name (%s)", tx.GetName())),
		})
	}
}

// handleRespTransaction passes the transaction to a response
// channel with a matching transaction id and deletes the channel afterwards.
func (c *Client) handleRespTransaction(tx *proto.Tx) error {
	if !txRespChannels.Has(tx.GetId()) {
		return fmt.Errorf("response transaction (%s) does not have a corresponding response channel", tx.GetId())
	}

	txRespCh, _ := txRespChannels.Get(tx.GetId())
	txRespCh.(chan *proto.Tx) <- tx
	txRespChannels.Remove(tx.GetId())
	return nil
}

// SendTx sends a transaction to the cocoon code
// and saves the response channel. The response channel will
// be passed a response when it is available in the Transact loop.
func (c *Client) SendTx(tx *proto.Tx, respCh chan *proto.Tx) error {
	txRespChannels.Set(tx.GetId(), respCh)
	if err := c.stream.Send(tx); err != nil {
		txRespChannels.Remove(tx.GetId())
		return err
	}
	log.Debugf("Successfully sent transaction (%s) to cocoon code", tx.GetId())
	return nil
}
