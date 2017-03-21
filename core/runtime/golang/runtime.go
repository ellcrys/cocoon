package golang

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/runtime/golang/config"
	"github.com/ncodes/cocoon/core/runtime/golang/proto"
	"github.com/ncodes/cocoon/core/types"
	"github.com/op/go-logging"
	cmap "github.com/orcaman/concurrent-map"
	"google.golang.org/grpc"
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
		if err = cc.Init(NewLink()); err != nil {
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
		Name:   types.TxPut,
		Params: []string{ledgerName},
		Body:   txsJSON,
	}, respCh)
	if err != nil {
		return fmt.Errorf("failed to put block transaction. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return err
	}

	if resp.Status != 200 {
		return fmt.Errorf("%s", common.GetRPCErrDesc(fmt.Errorf("%s", resp.Body)))
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
	if blockMaker != nil {
		blockMaker.Stop()
	}
	defaultServer.stream = nil
	serverDone <- true
	log.Info("Cocoon code exiting with exit code %d", exitCode)
	os.Exit(exitCode)
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

