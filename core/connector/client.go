package connector

import (
	"fmt"
	"os"
	"time"

	proto "github.com/ncodes/cocoon/core/runtime/golang/proto"
	logging "github.com/op/go-logging"
	cmap "github.com/orcaman/concurrent-map"
	context "golang.org/x/net/context"
)

var clientLog = logging.MustGetLogger("connector.client")

// txChannels holds the channels to send transaction responses to
var txRespChannels = cmap.New()

// Client represents a cocoon code GRPC client
// that interacts with a cocoon code.
type Client struct {
	ccodeAddr             string
	stub                  proto.StubClient
	conCtx                context.Context
	conCancel             context.CancelFunc
	orderDiscoTicker      *time.Ticker
	ordererAddrs          []string
	cocoonID              string
	streamKeepAliveTicker *time.Ticker
	stopped               bool
}

// NewClient creates a new cocoon code client
func NewClient() *Client {
	return &Client{}
}

// SetCocoonID sets the cocoon id
func (c *Client) SetCocoonID(id string) {
	c.cocoonID = id
}

// SetCocoonCodeAddr sets the cocoon code bind address
func (c *Client) SetCocoonCodeAddr(ccAddr string) {
	c.ccodeAddr = ccAddr
}

// getCCAddr returns the cocoon code bind address.
// For development, if DEV_COCOON_ADDR is set, it connects to it.
func (c *Client) getCCAddr() string {
	if devCCodeAddr := os.Getenv("DEV_COCOON_ADDR"); len(devCCodeAddr) > 0 {
		return devCCodeAddr
	}
	return c.ccodeAddr
}

// Close the stream and cancel connections
func (c *Client) Close() {
	c.stopped = true
	if c.orderDiscoTicker != nil {
		c.orderDiscoTicker.Stop()
	}
}

// Connect connects to a cocoon code server
// running on a known port
func (c *Client) Connect() error {

	// clientLog.Info("Starting cocoon code client")

	// // start a ticker to continously discover orderer addreses
	// go func() {
	// 	c.orderDiscoTicker = time.NewTicker(60 * time.Second)
	// 	for _ = range c.orderDiscoTicker.C {
	// 		var ordererAddrs []string
	// 		ordererAddrs, err := orderer.DiscoverOrderers()
	// 		if err != nil {
	// 			clientLog.Error(err.Error())
	// 			continue
	// 		}
	// 		c.ordererAddrs = ordererAddrs
	// 	}
	// }()

	// var ordererAddrs []string
	// ordererAddrs, err := orderer.DiscoverOrderers()
	// if err != nil {
	// 	return err
	// }
	// c.ordererAddrs = ordererAddrs

	// if len(c.ordererAddrs) > 0 {
	// 	clientLog.Infof("Orderer address list updated. Contains %d orderer address(es)", len(c.ordererAddrs))
	// } else {
	// 	clientLog.Warning("No orderer address was found. We won't be able to reach the orderer. ")
	// }

	// time.AfterFunc(2*time.Second, func() {
	// 	clientLog.Debugf("Now connected to cocoon code at port=%s", strings.Split(c.getCCAddr(), ":")[1])
	// })

	// if err = c.Do(); err != nil {
	// 	return err
	// }

	return nil
}

// Do starts a request loop that continously asks the
// server for transactions. When it receives a transaction,
// it processes it and returns the result.
func (c *Client) Do() error {
	return nil
	// var err error

	// conn, err := grpc.Dial(c.getCCAddr(), grpc.WithInsecure())
	// if err != nil {
	// 	return fmt.Errorf("failed to connect to cocoon code server. %s", err)
	// }
	// defer conn.Close()

	// c.stub = proto.NewStubClient(conn)

	// // create a context so we have complete controll of the connection
	// c.conCtx, c.conCancel = context.WithCancel(context.Background())

	// // connect to the cocoon code
	// c.stream, err = c.stub.Transact(c.conCtx)
	// if err != nil {
	// 	return fmt.Errorf("failed to start transaction stream with cocoon code. %s", err)
	// }

	// go c.keepStreamAlive()

	// for {
	// 	in, err := c.stream.Recv()
	// 	if err == io.EOF {
	// 		return fmt.Errorf("connection with cocoon code has ended")
	// 	}
	// 	if err != nil {
	// 		return fmt.Errorf("failed to read message from cocoon code. %s", err)
	// 	}

	// 	// keep alive message
	// 	if in.Invoke && in.Status == -100 {
	// 		clientLog.Debug("A keep alive message received")
	// 		continue
	// 	}

	// 	switch in.Invoke {
	// 	case true:
	// 		go func() {
	// 			clientLog.Debugf("New invoke transaction (%s) from cocoon code", in.GetId())
	// 			if err = c.handleInvokeTransaction(in); err != nil {
	// 				clientLog.Error(err.Error())
	// 				c.stream.Send(&proto.Tx{
	// 					Response: true,
	// 					Id:       in.GetId(),
	// 					Status:   500,
	// 					Body:     []byte(err.Error()),
	// 				})
	// 			}
	// 		}()
	// 	case false:
	// 		clientLog.Debugf("New response transaction (%s) from cocoon code", in.GetId())
	// 		go func() {
	// 			if err = c.handleRespTransaction(in); err != nil {
	// 				clientLog.Error(err.Error())
	// 				c.stream.Send(&proto.Tx{
	// 					Response: true,
	// 					Id:       in.GetId(),
	// 					Status:   500,
	// 					Body:     []byte(err.Error()),
	// 				})
	// 			}
	// 		}()
	// 	}
	// }
}

// handleInvokeTransaction processes invoke transaction requests
func (c *Client) handleInvokeTransaction(tx *proto.Tx) error {
	// switch tx.GetName() {
	// case types.TxCreateLedger:
	// 	return c.createLedger(tx)
	// case types.TxPut:
	// 	return c.put(tx)
	// case types.TxGetLedger:
	// 	return c.getLedger(tx)
	// case types.TxGet:
	// 	return c.get(tx, false)
	// case types.TxGetByID:
	// 	return c.get(tx, true)
	// case types.TxGetBlockByID:
	// 	return c.getBlock(tx)
	// case types.TxRangeGet:
	// 	return c.getRange(tx)
	// default:
	// 	return c.stream.Send(&proto.Tx{
	// 		Id:       tx.GetId(),
	// 		Response: true,
	// 		Status:   500,
	// 		Body:     []byte(fmt.Sprintf("unsupported transaction name (%s)", tx.GetName())),
	// 	})
	// }
	return nil
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
	// if c.stream != nil {
	// 	txRespChannels.Set(tx.GetId(), respCh)
	// 	if err := c.stream.Send(tx); err != nil {
	// 		txRespChannels.Remove(tx.GetId())
	// 		return err
	// 	}
	// 	clientLog.Debugf("Successfully sent transaction (%s) to cocoon code", tx.GetId())
	// 	return nil
	// }
	// return types.ErrUninitializedStream
	return nil
}
