package golang

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/stubs/golang/config"
	"github.com/ncodes/cocoon/core/stubs/golang/proto"
	"github.com/ncodes/cocoon/core/types"
	"github.com/op/go-logging"
	cmap "github.com/orcaman/concurrent-map"
	"google.golang.org/grpc"
)

const (
	// TxCreateLedger represents a message to create a ledger
	TxCreateLedger = "CREATE_LEDGER"

	// TxPut represents a message to create a transaction
	TxPut = "PUT"

	// TxGetLedger represents a message to get a ledger
	TxGetLedger = "GET_LEDGER"

	// TxGet represents a message to get a transaction
	TxGet = "GET"

	// TxGetByID represents a message to get a transaction by id
	TxGetByID = "GET_BY_ID"
)

var (

	// serverPort to bind to
	serverPort = util.Env("COCOON_CODE_PORT", "8000")

	// stub logger
	log *logging.Logger

	// default running server
	defaultServer *stubServer

	// stop channel to stop the server/cocoon code
	serverDone chan bool

	// The default ledger is the global ledger.
	defaultLedger = GetGlobalLedgerName()

	// txChannels holds the channels to send transaction responses to
	txRespChannels = cmap.New()

	// ErrOperationTimeout represents a timeout error that occurs when response
	// is not received from orderer in time.
	ErrOperationTimeout = fmt.Errorf("operation timed out")

	// ErrAlreadyExist represents an error about an already existing resource
	ErrAlreadyExist = fmt.Errorf("already exists")

	// ErrNotConnected represents an error about the cocoon code not
	// having an active connection with the connector.
	ErrNotConnected = fmt.Errorf("not connected to the connector")

	// Flag to help tell whether cocoon code is running
	running = false

	// Number of transactions per block
	txPerBlock = util.Env("TX_PER_BLOCK", "100")

	// Time between block creation (seconds)
	blockCreationInterval = util.Env("BLOCK_CREATION_INT", "5")

	// blockMaker creates a collection of blockchain transactions at interval
	blockMaker *BlockMaker

	// The cocoon code currently running
	ccode CocoonCode
)

// GetLogger returns the stubs logger.
func GetLogger() *logging.Logger {
	return log
}

// SetDebugLevel sets the default logger debug level
func SetDebugLevel(level logging.Level) {
	logging.SetLevel(level, log.Module)
}

func init() {
	defaultServer = new(stubServer)
	config.ConfigureLogger()
	log = logging.MustGetLogger("ccode.stub")
}

// GetGlobalLedgerName returns the name of the global ledger
func GetGlobalLedgerName() string {
	return types.GetGlobalLedgerName()
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

	intTxPerBlock, _ := strconv.Atoi(txPerBlock)
	intBlkCreationInt, _ := strconv.Atoi(blockCreationInterval)
	blockMaker = NewBlockMaker(intTxPerBlock, time.Duration(intBlkCreationInt)*time.Second)
	go blockMaker.Begin(blockCommitter)

	ccode = cc

	// run Init() after 1 second to give time for connector to connect
	time.AfterFunc(1*time.Second, func() {
		if err = cc.Init(); err != nil {
			log.Errorf("cocoode Init() returned error: %s", err)
			Stop(1)
		} else {
			running = true
		}
	})

	<-serverDone
	log.Info("Cocoon code stopped")
	os.Exit(0)
}

// blockCommit creates a PUT operation which adds one or many
// transactions to the store and blockchain and returns the block if
// if succeed or error if otherwise.
func blockCommitter(entries []*Entry) interface{} {

	txs := make([]*types.Transaction, len(entries))
	for i, e := range entries {
		txs[i] = e.Tx
	}

	ledgerName := entries[0].Tx.Ledger
	txsJSON, _ := util.ToJSON(txs)

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   TxPut,
		Params: []string{ledgerName},
		Body:   txsJSON,
	}, respCh)
	if err != nil {
		return fmt.Errorf("failed to put block transaction. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return err
	}

	if resp.Status != 200 {
		return fmt.Errorf("%s", common.StripRPCErrorPrefix(resp.Body))
	}

	var block types.Block
	if err = util.FromJSON(resp.Body, &block); err != nil {
		return fmt.Errorf("failed to unmarshall response data")
	}

	return &block
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

// SetDefaultLedger sets the default ledger
func SetDefaultLedger(name string) error {
	_, err := GetLedger(name)
	if err != nil {
		return err
	}
	defaultLedger = name
	return nil
}

// GetDefaultLedgerName returns the name of the default ledger.
func GetDefaultLedgerName() string {
	return defaultLedger
}

// CreateLedger creates a new ledger by sending an
// invoke transaction (TxCreateLedger) to the connector.
// If chained is set to true, a blockchain is created and subsequent
// PUT operations to the ledger will be included in the types. Otherwise,
// PUT operations will only be incuded in the types.
func CreateLedger(name string, chained, public bool) (*types.Ledger, error) {

	if name == GetGlobalLedgerName() {
		return nil, fmt.Errorf("cannot use the same name as the global ledger")
	} else if !common.IsValidResName(name) {
		return nil, fmt.Errorf("invalid ledger name")
	}

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   TxCreateLedger,
		Params: []string{name, fmt.Sprintf("%t", chained), fmt.Sprintf("%t", public)},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to create ledger. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}

	if resp.Status != 200 {
		err = fmt.Errorf("%s", common.StripRPCErrorPrefix(resp.Body))
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
		return nil, fmt.Errorf("%s", common.StripRPCErrorPrefix(resp.Body))
	}

	var ledger types.Ledger
	if err = util.FromJSON(resp.Body, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &ledger, nil
}

// PutIn adds a new transaction to a ledger
func PutIn(ledgerName string, key string, value []byte) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	ledger, err := GetLedger(ledgerName)
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ID:        util.UUID4(),
		Ledger:    ledger.Name,
		Key:       key,
		Value:     string(value),
		CreatedAt: time.Now().Unix(),
	}
	tx.Hash = tx.MakeHash()

	if ledger.Chained {
		respChan := make(chan interface{})
		blockMaker.Add(&Entry{
			Tx:       tx,
			RespChan: respChan,
		})
		result := <-respChan

		switch v := result.(type) {
		case error:
			return nil, v
		case *types.Block:
			tx.Block = v
			return tx, err
		default:
			return nil, fmt.Errorf("unexpected response %s", err)
		}
	}

	txJSON, _ := util.ToJSON([]*types.Transaction{tx})

	var respCh = make(chan *proto.Tx)
	err = sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		Name:   TxPut,
		Params: []string{ledgerName},
		Body:   txJSON,
	}, respCh)

	if err != nil {
		return nil, fmt.Errorf("failed to put transaction. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", common.StripRPCErrorPrefix(resp.Body))
	}

	return tx, nil
}

// Put adds a new transaction into the default ledger
func Put(key string, value []byte) (*types.Transaction, error) {
	return PutIn(GetDefaultLedgerName(), key, value)
}

// GetFrom returns a transaction by its key and the ledger it belongs to
func GetFrom(ledgerName, key string) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
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
		return nil, fmt.Errorf("%s", common.StripRPCErrorPrefix(resp.Body))
	}

	var tx types.Transaction
	if err = util.FromJSON(resp.Body, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &tx, nil
}

// Get returns a transaction that belongs to the default legder by its key.
func Get(key string) (*types.Transaction, error) {
	return GetFrom(GetDefaultLedgerName(), key)
}

// GetByIDFrom returns a transaction by its id and the ledger it belongs to
func GetByIDFrom(ledgerName, id string) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		Name:   TxGetByID,
		Params: []string{ledgerName, id},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	}

	resp, err := AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", common.StripRPCErrorPrefix(resp.Body))
	}

	var tx types.Transaction
	if err = util.FromJSON(resp.Body, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &tx, nil
}

// GetByID returns a transaction that belongs to the default legder by its id.
func GetByID(id string) (*types.Transaction, error) {
	return GetByIDFrom(GetDefaultLedgerName(), id)
}
