package golang

import (
	"fmt"
	"io"
	"net"

	"os"

	"time"

	"strings"

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

const (
	// TxCreateLedger represents a message to create a ledger
	TxCreateLedger = "CREATE_LEDGER"

	// TxPut represents a message to create a transaction
	TxPut = "PUT"

	// TxGetLedger represents a message to get a ledger
	TxGetLedger = "GET_LEDGER"

	// TxGet represents a message to get a transaction
	TxGet = "GET"
)

// The default ledger is the global ledger.
var (
	defaultLedger = GetGlobalLedgerName()

	// txChannels holds the channels to send transaction responses to
	txRespChannels = cmap.New()

	// ErrOperationTimeout represents a timeout error that occurs when response
	// is not received from orderer in time.
	ErrOperationTimeout = fmt.Errorf("operation timed out")

	// ErrNotFound represents an error about a resource not found
	ErrNotFound = fmt.Errorf("not found")

	// ErrAlreadyExist represents an error about an already existing resource
	ErrAlreadyExist = fmt.Errorf("already exists")

	// ErrNotConnected represents an error about the cocoon code not
	// having an active connection with the connector.
	ErrNotConnected = fmt.Errorf("not connected to the connector")

	// Flag to help tell whether cocoon code is running
	running = false

	// The cocoon code currently running
	ccode CocoonCode
)

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

// GetGlobalLedgerName returns the name of the global ledger
func GetGlobalLedgerName() string {
	return types.GetGlobalLedgerName()
}

// stripRPCErrorPrefix takes an error return from the RPC client and removes the
// prefixed `rpc error: code = 2 desc =` statement
func stripRPCErrorPrefix(err []byte) []byte {
	return []byte(strings.TrimSpace(strings.Replace(string(err), "rpc error: code = 2 desc =", "", -1)))
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
	switch tx.GetName() {
	case "function":
		if !running {
			return fmt.Errorf("cocoon code is not running. Did you call the Run() method?")
		}

		functionName := tx.GetParams()[0]
		result, err := ccode.Invoke(tx.GetId(), functionName, tx.GetParams()[1:])
		if err != nil {
			return err
		}

		// coerce result to json
		resultJSON, err := util.ToJSON(result)
		if err != nil {
			return fmt.Errorf("failed to coerce cocoon code Invoke() result to json string. %s", err)
		}

		return s.stream.Send(&proto.Tx{
			Id:       tx.GetId(),
			Response: true,
			Status:   200,
			Body:     resultJSON,
		})

	default:
		return fmt.Errorf("Unsupported invoke transaction (%s)", tx.GetName())
	}
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

// Run starts the stub server, takes a cocoon code and attempts to initialize it..
func Run(cc CocoonCode) {

	if running {
		log.Info("cocoon code is already running")
		return
	}

	serverDone = make(chan bool, 1)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		log.Fatalf("failed to listen on port=%s", serverPort)
	}

	log.Infof("Started stub service at port=%s", serverPort)
	server := grpc.NewServer()
	proto.RegisterStubServer(server, defaultServer)
	go server.Serve(lis)

	if err = cc.Init(); err != nil {
		log.Errorf("cocoode Init() returned error: %s", err)
		Stop(1)
	}

	running = true
	ccode = cc

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
		log.Errorf("failed to send transaction [%s] to connector. %s", tx.GetId(), err)
		return err
	}
	log.Debugf("Successfully sent transaction [%s] to connector", tx.GetId())
	return nil
}

// Stop stub and cocoon code
func Stop(exitCode int) {
	defaultServer.stream = nil
	serverDone <- true
	log.Info("Cocoon code exiting with exit code %d", exitCode)
	os.Exit(exitCode)
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

// SetDefault ledger
func SetDefault(name string) {
	// ledger, err := GetLedger(name)
	defaultLedger = name
}

// GetDefaultLedger returns the name of the default ledger.
func GetDefaultLedger() string {
	return defaultLedger
}

// CreateLedger creates a new ledger by sending an
// invoke transaction (TxCreateLedger) to the connector.
// The final name of the ledger is a sha256 hash of
// the cocoon code id and the name (e.g SHA256(ccode_id.name))
func CreateLedger(name string, public bool) (*types.Ledger, error) {

	// TODO: prevent use of ledger name with punctuations (execept an underscore)

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   TxCreateLedger,
		Params: []string{name, fmt.Sprintf("%t", public)},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to create ledger. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}

	if resp.Status != 200 {
		err = fmt.Errorf("%s", stripRPCErrorPrefix(resp.Body))
		if strings.Contains(err.Error(), "already exists") {
			return nil, ErrAlreadyExist
		}
		return nil, err
	}

	var ledger types.Ledger
	if err = util.FromJSON(resp.Body, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &ledger, nil
}

// GetLedger fetches a ledger
func GetLedger(ledgerName string) (*types.Ledger, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   TxGetLedger,
		Params: []string{ledgerName},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", stripRPCErrorPrefix(resp.Body))
	}

	var ledger types.Ledger
	if err = util.FromJSON(resp.Body, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	if err == nil && ledger == (types.Ledger{}) {
		return nil, ErrNotFound
	}

	return &ledger, nil
}

// PutIn adds a new transaction to a ledger
func PutIn(ledgerName string, key string, value []byte) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   TxPut,
		Params: []string{ledgerName, txID, key, string(value)},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to put transaction. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", stripRPCErrorPrefix(resp.Body))
	}

	var tx types.Transaction
	if err = util.FromJSON(resp.Body, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &tx, nil
}

// Put adds a new transaction into the default ledger
func Put(key string, value []byte) (*types.Transaction, error) {
	return PutIn(GetDefaultLedger(), key, value)
}

// GetFrom returns a transaction by its key and the ledger it belongs to
func GetFrom(ledgerName, key string) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   TxGet,
		Params: []string{ledgerName, key},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", stripRPCErrorPrefix(resp.Body))
	}

	var tx types.Transaction
	if err = util.FromJSON(resp.Body, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &tx, nil
}

// Get returns a transaction that belongs to the default legder by its key.
func Get(key string) (*types.Transaction, error) {
	return GetFrom(GetDefaultLedger(), key)
}
