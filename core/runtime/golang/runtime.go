// Package golang includes The runtime, which is just a stub that connects the cocoon code to the connector's RPC
// server. The runtime provides access to APIs for various operations.
package golang

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"strings"

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

	// serverAddr to bind to
	serverAddr = util.Env("COCOON_ADDR", ":8000")

	// stub logger
	log *logging.Logger

	// default running server
	defaultServer *stubServer

	// stop channel to stop the server/cocoon code
	serverDone chan bool

	// Default runtime link
	defaultLink = NewNativeLink(GetID())

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
	log = logging.MustGetLogger("ccode.runtime")
}

// GetID returns the cocoon id. However, it will return the
// natively linked cocoon id if this cocoon is linked to another
// cocoon.
func GetID() string {
	return util.Env("COCOON_LINK", os.Getenv("COCOON_ID"))
}

// GetCocoonID returns the unique cocoon id
func GetCocoonID() string {
	return os.Getenv("COCOON_ID")
}

// Run starts the stub server, takes a cocoon code and attempts to initialize it..
func Run(cc CocoonCode) {

	if running {
		log.Info("cocoon code is already running")
		return
	}

	serverDone = make(chan bool, 1)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", serverAddr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s", strings.Split(serverAddr, ":")[1])
	}

	log.Infof("Started stub service at port=%s", strings.Split(serverAddr, ":")[1])
	server := grpc.NewServer()
	proto.RegisterStubServer(server, defaultServer)
	go startServer(server, lis)

	intTxPerBlock, _ := strconv.Atoi(txPerBlock)
	intBlkCreationInt, _ := strconv.Atoi(blockCreationInterval)
	blockMaker = NewBlockMaker(intTxPerBlock, time.Duration(intBlkCreationInt)*time.Second)
	go blockMaker.Begin(blockCommitter)

	defaultLink.SetDefaultLedger(types.GetGlobalLedgerName())

	ccode = cc

	// run Init() after 1 second to give time for connector to connect
	time.AfterFunc(1*time.Second, func() {
		if err = cc.OnInit(defaultLink); err != nil {
			log.Errorf("cocoode OnInit() returned error: %s", err)
			Stop(2)
		} else {
			running = true
		}
	})

	<-serverDone
	log.Info("Cocoon code stopped")
	os.Exit(0)
}

// startServer starts the server
func startServer(server *grpc.Server, lis net.Listener) {
	err := server.Serve(lis)
	if err != nil {
		log.Errorf("Server stopped: %s", err)
		Stop(1)
	}
}

// blockCommit creates a PUT operation which adds one or many
// transactions to the store and blockchain and returns the block if
// if succeed or error if otherwise.
func blockCommitter(entries []*Entry) interface{} {

	var block types.Block
	var respCh = make(chan *proto.Tx)

	if len(entries) == 0 {
		return block
	}

	txs := make([]*types.Transaction, len(entries))
	for i, e := range entries {
		txs[i] = e.Tx
	}

	ledgerName := entries[0].Tx.Ledger
	txsJSON, _ := util.ToJSON(txs)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		Name:   types.TxPut,
		LinkTo: entries[0].To,
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
	if defaultServer.streamKeepAliveTicker != nil {
		defaultServer.streamKeepAliveTicker.Stop()
	}
	defaultServer.stream = nil
	serverDone <- true
	running = false
	log.Info("Cocoon code exiting with exit code %d", exitCode)
	os.Exit(exitCode)
}

// isConnected checks if connection with the connector
// is active.
func isConnected() bool {
	return defaultServer.stream != nil
}

// GetGlobalLedger returns the name of the global ledger.
func GetGlobalLedger() string {
	return types.GetGlobalLedgerName()
}
