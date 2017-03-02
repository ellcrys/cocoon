package golang

import (
	"fmt"
	"io"
	"net"

	"os"

	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/ledgerchain/types"
	"github.com/ncodes/cocoon/core/stubs/golang/config"
	"github.com/ncodes/cocoon/core/stubs/golang/proto"
	"github.com/op/go-logging"
	cmap "github.com/orcaman/concurrent-map"
	"google.golang.org/grpc"
)

var serverPort = util.Env("COCOON_CODE_PORT", "8000")
var defaultServer *stubServer
var log *logging.Logger
var serverDone chan bool

// txChannels holds the channels to send transaction responses to
var txRespChannels = cmap.New()

// ErrOperationTimeout represents a timeout error that occurs when response
// is not received from orderer in time.
var ErrOperationTimeout = fmt.Errorf("operation timed out")

// ErrNotConnected represents an error about the cocoon code not
// having an active connection with the connector.
var ErrNotConnected = fmt.Errorf("not connected to the connector")

// TxListLedgers represents a message to list ledgers belonging to cocoon code
const TxListLedgers = "LIST_LEDGERS"

// TxCreateLedger represents a message to create a ledger
const TxCreateLedger = "CREATE_LEDGER"

func init() {
	defaultServer = new(stubServer)
	config.ConfigureLogger()
	log = logging.MustGetLogger("ccode.stub")
}

// StubServer defines the services of the stub's GRPC connection
type stubServer struct {
	port   int
	stream proto.Stub_TransactServer
}

// Transact listens and process invoke and response transactions from
// the connector.
func (s *stubServer) Transact(stream proto.Stub_TransactServer) error {
	s.stream = stream
	for {

		in, err := stream.Recv()
		if err == io.EOF {
			return fmt.Errorf("connection with cocoon code has ended")
		}
		if err != nil {
			return fmt.Errorf("failed to read message from connector. %s", err)
		}

		switch in.Invoke {
		case true:
			go func() {
				log.Debugf("New invoke transaction (%s) from connector", in.GetId())
				if err = s.handleInvokeTransaction(in); err != nil {
					log.Error(err.Error())
					stream.Send(&proto.Tx{
						Response: true,
						Id:       in.GetId(),
						Status:   500,
						Body:     []byte(err.Error()),
					})
				}
			}()
		case false:
		default:
			log.Debugf("New response transaction (%s) from connector", in.GetId())
			go func() {
				if err = s.handleRespTransaction(in); err != nil {
					log.Error(err.Error())
					stream.Send(&proto.Tx{
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
func (s *stubServer) handleInvokeTransaction(tx *proto.Tx) error {
	return s.stream.Send(&proto.Tx{
		Id:       tx.GetId(),
		Response: true,
	})
}

// handleRespTransaction passes the transaction to a response
// channel with a matching transaction id and deletes the channel afterwards.
func (s *stubServer) handleRespTransaction(tx *proto.Tx) error {
	if !txRespChannels.Has(tx.GetId()) {
		return fmt.Errorf("response transaction (%s) does not have a corresponding response channel", tx.GetId())
	}

	txRespCh, _ := txRespChannels.Get(tx.GetId())
	txRespCh.(chan *proto.Tx) <- tx
	txRespChannels.Remove(tx.GetId())
	return nil
}

// StartServer starts the stub server and
// readys it for service processing.
// Accepts a callback that is called when the server starts
func StartServer(startedCb func()) {

	serverDone = make(chan bool, 1)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		log.Fatalf("failed to listen on port=%s", serverPort)
	}

	log.Infof("Started stub service at port=%s", serverPort)
	server := grpc.NewServer()
	proto.RegisterStubServer(server, &stubServer{})
	go server.Serve(lis)

	startedCb()
	<-serverDone
	log.Info("Cocoon code stopped")
	os.Exit(0)
}

// GetLogger returns the stubs logger.
func GetLogger() *logging.Logger {
	return log
}

// sendTx sends a transaction to the cocoon code
// and saves the response channel. The response channel will
// be passed a response when it is available in the Transact loop.
func sendTx(tx *proto.Tx, respCh chan *proto.Tx) error {
	txRespChannels.Set(tx.GetId(), respCh)
	if err := defaultServer.stream.Send(tx); err != nil {
		txRespChannels.Remove(tx.GetId())
		log.Debugf("Failed to send transaction [%s] to orderer. %s", tx.GetId(), err)
		return err
	}
	log.Debugf("Successfully sent transaction [%s] to orderer", tx.GetId())
	return nil
}

// Stop stub and cocoon code
func Stop() {
	log.Info("Stopping cocoon code")
	defaultServer.stream = nil
	serverDone <- true
}

// AwaitTxChan takes a response channel and waits to receive a response
// from it. If no error occurs, it returns the response. It
// returns ErrOperationTimeout if it waited 5 minutes and got no response.
func AwaitTxChan(ch chan *proto.Tx) (*proto.Tx, error) {
	for {
		select {
		case r := <-ch:
			return r, nil
		case <-time.After(5 * time.Minute):
			return nil, ErrOperationTimeout
		}
	}
}

// isConnected checks if connection with the connector
// is active.
func isConnected() bool {
	return defaultServer.stream != nil
}

// ListLedgers returns the list of ledgers created
// by the current cocoon code.
func ListLedgers() ([]*types.Ledger, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Name:   TxListLedgers,
		Params: []string{},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed get ledger list. %s", err)
	}

	// wait for response
	_, err = AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// CreateLedger creates a new ledger
func CreateLedger() (*types.Ledger, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Name:   TxCreateLedger,
		Params: []string{},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed create ledger. %s", err)
	}

	log.Debug("Waiting for response for transaction %s", txID)
	resp, err := AwaitTxChan(respCh)
	if err != nil {
		log.Errorf("receiving message from transaction [%s] failed because: %s", txID, err)
		return nil, err
	}

	log.Info("Got Response: %s", resp)
	return nil, nil
}
