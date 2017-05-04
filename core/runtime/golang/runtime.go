// Package golang includes The runtime, which is just a stub that connects the cocoon code to the connector's RPC
// server. The runtime provides access to APIs for various operations.
package golang

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	"github.com/ncodes/cocoon/core/runtime/golang/config"
	"github.com/ncodes/cocoon/core/runtime/golang/proto_runtime"
	"github.com/ncodes/cocoon/core/types"
	"github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (

	// serverAddr to bind to
	serverAddr = util.Env("COCOON_RPC_ADDR", ":8000")

	// connector's RPC address
	connectorRPCAddr = os.Getenv("CONNECTOR_RPC_ADDR")

	// stub logger
	log *logging.Logger

	// default running server
	defaultServer *stubServer

	// stop channel to stop the server/cocoon code
	serverDone chan bool

	// System link
	System = NewLink(types.SystemCocoonID)

	// Native link refers to a link that points
	// to the resources of a natively linked cocoon. If the current cocoon
	// has not native link to another cocoon, then the link points to its own resource
	Native = newNativeLink(GetID())

	// Me link refers to a link that points the resources of the current cocoon respective
	// of whether the cocoon is natively linked to another cocoon.
	Me = NewLink(GetCocoonID())

	// Flag to help tell whether cocoon code is running
	running = false

	// Number of transactions per block
	txPerBlock = util.Env("TX_PER_BLOCK", "100")

	// Time between block creation (seconds)
	blockCreationInterval = util.Env("BLOCK_CREATION_INT", "5")

	// defaultBlockMaker creates a collection of blockchain transactions at interval
	defaultBlockMaker *blockMaker

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

// GetSystemCocoonID returns the system cocoon id
func GetSystemCocoonID() string {
	return types.SystemCocoonID
}

// Run starts the stub server, takes a cocoon code and attempts to initialize it..
func Run(cc CocoonCode) {

	if running {
		log.Info("cocoon code is already running")
		return
	}

	serverDone = make(chan bool, 1)

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s", serverAddr)
	}

	log.Infof("Started stub service started %s", serverAddr)
	server := grpc.NewServer()
	proto_runtime.RegisterStubServer(server, defaultServer)
	go startServer(server, lis)

	intTxPerBlock, _ := strconv.Atoi(txPerBlock)
	intBlkCreationInt, _ := strconv.Atoi(blockCreationInterval)
	defaultBlockMaker = newblockMaker(intTxPerBlock, time.Duration(intBlkCreationInt)*time.Second)
	go defaultBlockMaker.Begin(blockCommitter)

	ccode = cc

	// run Init() after 1 second to give time for connector to connect
	time.AfterFunc(1*time.Second, func() {
		if err = cc.OnInit(); err != nil {
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
		log.Errorf("server has stopped: %s", err)
		Stop(1)
	}
}

// blockCommit creates a PUT operation which adds one or many
// transactions to the store and blockchain and returns the block if
// if succeed or error if otherwise.
func blockCommitter(entries []*entry) interface{} {

	var putResult types.PutResult
	if len(entries) == 0 {
		return fmt.Errorf("empty entry list")
	}

	txs := make([]*types.Transaction, len(entries))
	for i, e := range entries {
		txs[i] = e.Tx
	}
	ledgerName := entries[0].Tx.Ledger
	txsJSON, _ := util.ToJSON(txs)
	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxPut,
		LinkTo: entries[0].LinkTo,
		Params: []string{ledgerName},
		Body:   txsJSON,
	})

	if err != nil {
		return fmt.Errorf("failed to put block transaction: %s", err)
	}

	if err = util.FromJSON(result, &putResult); err != nil {
		return fmt.Errorf("failed to unmarshal response data")
	}

	return &putResult
}

// sendOp sends out a request to the connector.
func sendOp(req *proto_connector.Request) ([]byte, error) {

	client, err := grpc.Dial(connectorRPCAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ccClient := proto_connector.NewConnectorClient(client)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)
	resp, err := ccClient.Transact(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s", common.GetRPCErrDesc(err))
	}

	if resp.GetStatus() == 500 {
		return nil, fmt.Errorf("server error")
	}

	return resp.GetBody(), nil
}

// sendLedgerOp sends a ledger transaction to the connector
func sendLedgerOp(op *proto_connector.LedgerOperation) ([]byte, error) {
	return sendOp(&proto_connector.Request{
		OpType:   proto_connector.OpType_LedgerOp,
		LedgerOp: op,
	})
}

// sendLockOp sends a lock operation to the connector
func sendLockOp(op *proto_connector.LockOperation) ([]byte, error) {
	return sendOp(&proto_connector.Request{
		OpType: proto_connector.OpType_LockOp,
		LockOp: op,
	})
}

// Stop stub and cocoon code
func Stop(exitCode int) {
	if defaultBlockMaker != nil {
		defaultBlockMaker.Stop()
	}

	serverDone <- true
	running = false
	log.Info("Cocoon code exiting with exit code %d", exitCode)
	os.Exit(exitCode)
}

// GetSystemPublicLedgerName returns the name of the system ledger.
func GetSystemPublicLedgerName() string {
	return types.GetSystemPublicLedgerName()
}
