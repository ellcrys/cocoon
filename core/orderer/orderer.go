package orderer

import (
	"fmt"
	"net"
	"strings"

	context "golang.org/x/net/context"

	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/ledgerchain/types"
	"github.com/ncodes/cocoon/core/orderer/proto"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("orderer")

// Orderer defines a transaction ordering, block creation
// and inclusion module
type Orderer struct {
	server  *grpc.Server
	chain   types.LedgerChain
	endedCh chan bool
}

// NewOrderer creates a new Orderer object
func NewOrderer() *Orderer {
	return new(Orderer)
}

// Start starts the order service
func (od *Orderer) Start(addr, ledgerChainConStr string, endedCh chan bool) {

	od.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	time.AfterFunc(2*time.Second, func() {

		log.Infof("Started orderer GRPC server on port %s", strings.Split(addr, ":")[1])

		// establish connection to chain backend
		_, err := od.chain.Connect(ledgerChainConStr)
		if err != nil {
			log.Info(err)
			od.Stop(1)
			return
		}

		// initialize ledgerchain
		err = od.chain.Init(od.chain.MakeLedgerName("", types.GetGlobalLedgerName()))
		if err != nil {
			log.Info(err)
			od.Stop(1)
			return
		}

		log.Info("Backend successfully connnected")
	})

	od.server = grpc.NewServer()
	proto.RegisterOrdererServer(od.server, od)
	od.server.Serve(lis)
}

// Stop stops the orderer and returns an exit code.
func (od *Orderer) Stop(exitCode int) int {
	od.server.Stop()
	od.chain.Close()
	close(od.endedCh)
	return exitCode
}

// SetLedgerChain sets the ledgerchain implementation to use.
func (od *Orderer) SetLedgerChain(ch types.LedgerChain) {
	log.Infof("Setting ledgerchain backend to %s", ch.GetBackend())
	od.chain = ch
}

// CreateLedger creates a new ledger
func (od *Orderer) CreateLedger(ctx context.Context, params *proto.CreateLedgerParams) (*proto.Ledger, error) {

	name := od.chain.MakeLedgerName(params.GetCocoonCodeId(), params.GetName())
	ledger, err := od.chain.CreateLedger(name, params.GetPublic())
	if err != nil {
		return nil, err
	}

	ledgerJSON, _ := util.ToJSON(ledger)
	var protoLedger proto.Ledger
	util.FromJSON(ledgerJSON, &protoLedger)

	return &protoLedger, nil
}

// GetLedger returns a ledger
func (od *Orderer) GetLedger(ctx context.Context, params *proto.GetLedgerParams) (*proto.Ledger, error) {

	name := od.chain.MakeLedgerName(params.GetCocoonCodeId(), params.GetName())
	ledger, err := od.chain.GetLedger(name)
	if err != nil {
		return nil, err
	}

	ledgerJSON, _ := util.ToJSON(ledger)
	var protoLedger proto.Ledger
	util.FromJSON(ledgerJSON, &protoLedger)

	return &protoLedger, nil
}

// Put creates a new transaction
func (od *Orderer) Put(ctx context.Context, params *proto.PutTransactionParams) (*proto.Transaction, error) {

	// check if ledger exists
	name := od.chain.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedgerName())
	ledger, err := od.chain.GetLedger(name)
	if err != nil {
		return nil, err
	}
	if err == nil && ledger == nil {
		return nil, fmt.Errorf("ledger not found")
	}

	key := od.chain.MakeTxKey(params.GetCocoonCodeId(), params.GetKey())
	tx, err := od.chain.Put(params.GetId(), params.GetLedgerName(), key, string(params.GetValue()))
	if err != nil {
		return nil, err
	}

	txJSON, _ := util.ToJSON(tx)
	var protoTx proto.Transaction
	util.FromJSON(txJSON, &protoTx)

	return &protoTx, nil
}

// Get returns a transaction with a matching key and cocoon id
func (od *Orderer) Get(ctx context.Context, params *proto.GetParams) (*proto.Transaction, error) {

	key := od.chain.MakeTxKey(params.GetCocoonCodeId(), params.GetKey())
	tx, err := od.chain.Get(params.GetLedger(), key)
	if err != nil {
		return nil, err
	}

	txJSON, _ := util.ToJSON(tx)
	var protoTx proto.Transaction
	util.FromJSON(txJSON, &protoTx)

	return &protoTx, nil
}
