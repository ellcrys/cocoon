package orderer

import (
	"fmt"
	"net"
	"os"
	"strings"

	context "golang.org/x/net/context"

	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cocoon/core/types/store"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("orderer")

// DiscoverOrderers fetches a list of orderer service addresses
// via consul service discovery API. For development purpose,
// If DEV_ORDERER_ADDR is set, it will fetch the orderer
// address from the env variable.
func DiscoverOrderers() []string {
	if len(os.Getenv("DEV_ORDERER_ADDR")) > 0 {
		return []string{os.Getenv("DEV_ORDERER_ADDR")}
	}
	// TODO: Retrieve from consul service API (not implemented)
	return []string{}
}

// DialOrderer returns a connection to a orderer from a list of addresses. It randomly
// picks an orderer address from the list for orderers.
func DialOrderer(orderersAddr []string) (*grpc.ClientConn, error) {
	var ordererAddr string

	if len(orderersAddr) == 0 {
		return nil, fmt.Errorf("no known orderer address")
	} else if len(orderersAddr) == 1 {
		ordererAddr = orderersAddr[0]
	} else {
		ordererAddr = orderersAddr[util.RandNum(0, len(orderersAddr))]
	}

	client, err := grpc.Dial(ordererAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Orderer defines a transaction ordering, block creation
// and inclusion module
type Orderer struct {
	server  *grpc.Server
	chain   store.Store
	endedCh chan bool
}

// NewOrderer creates a new Orderer object
func NewOrderer() *Orderer {
	return new(Orderer)
}

// Start starts the order service
func (od *Orderer) Start(addr, storeConStr string, endedCh chan bool) {

	od.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	time.AfterFunc(2*time.Second, func() {

		log.Infof("Started orderer GRPC server on port %s", strings.Split(addr, ":")[1])

		// establish connection to chain backend
		_, err := od.chain.Connect(storeConStr)
		if err != nil {
			log.Info(err)
			od.Stop(1)
			return
		}

		// initialize store
		err = od.chain.Init(od.chain.MakeLedgerName("", store.GetGlobalLedgerName()))
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

// SetStore sets the store implementation to use.
func (od *Orderer) SetStore(ch store.Store) {
	log.Infof("Setting store backend to %s", ch.GetBackend())
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
	} else if ledger == nil && err == nil {
		return nil, types.ErrLedgerNotFound
	}

	ledgerJSON, _ := util.ToJSON(ledger)
	var protoLedger proto.Ledger
	util.FromJSON(ledgerJSON, &protoLedger)

	return &protoLedger, nil
}

// Put creates a new transaction
func (od *Orderer) Put(ctx context.Context, params *proto.PutTransactionParams) (*proto.Transaction, error) {

	// check if ledger exists
	ledgerName := od.chain.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedgerName())
	ledger, err := od.chain.GetLedger(ledgerName)
	if err != nil {
		return nil, err
	} else if err == nil && ledger == nil {
		return nil, types.ErrLedgerNotFound
	}

	key := od.chain.MakeTxKey(params.GetCocoonCodeId(), params.GetKey())
	tx, err := od.chain.Put(params.GetId(), ledgerName, key, string(params.GetValue()))
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

	_, err := od.GetLedger(ctx, &proto.GetLedgerParams{
		CocoonCodeId: params.GetCocoonCodeId(),
		Name:         params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	ledgerName := od.chain.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedger())
	key := od.chain.MakeTxKey(params.GetCocoonCodeId(), params.GetKey())
	tx, err := od.chain.Get(ledgerName, key)
	if err != nil {
		return nil, err
	} else if tx == nil && err == nil {
		return nil, types.ErrTxNotFound
	}

	txJSON, _ := util.ToJSON(tx)
	var protoTx proto.Transaction
	util.FromJSON(txJSON, &protoTx)

	return &protoTx, nil
}
