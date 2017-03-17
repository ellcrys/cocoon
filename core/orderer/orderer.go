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
	server     *grpc.Server
	store      types.Store
	blockchain types.Blockchain
	endedCh    chan bool
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

		// establish connection to store backend
		_, err := od.store.Connect(storeConStr)
		if err != nil {
			log.Info(err)
			od.Stop(1)
			return
		}

		// initialize store
		if od.store == nil {
			log.Error("Store implementation not set")
			od.Stop(1)
			return
		}

		err = od.store.Init(od.store.MakeLedgerName("", types.GetGlobalLedgerName()))
		if err != nil {
			log.Info(err)
			od.Stop(1)
			return
		}

		if od.blockchain == nil {
			log.Error("Blockchain implementation not set")
			od.Stop(1)
			return
		}

		// establish connection to blockchain backend
		if od.blockchain != nil {
			_, err = od.blockchain.Connect(storeConStr)
			if err != nil {
				log.Info(err)
				od.Stop(1)
				return
			}
		}

		// initialize the blockchain
		err = od.blockchain.Init(od.blockchain.MakeChainName("", types.GetGlobalChainName()))
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
	od.store.Close()
	close(od.endedCh)
	return exitCode
}

// SetStore sets the store implementation to use.
func (od *Orderer) SetStore(ch types.Store) {
	log.Infof("Setting store implementation named %s", ch.GetImplmentationName())
	od.store = ch
}

// SetBlockchain sets the blockchain implementation
func (od *Orderer) SetBlockchain(b types.Blockchain) {
	log.Infof("Setting blockchain implementation named %s", b.GetImplmentationName())
	od.blockchain = b
}

// CreateLedger creates a new ledger
func (od *Orderer) CreateLedger(ctx context.Context, params *proto.CreateLedgerParams) (*proto.Ledger, error) {

	name := od.store.MakeLedgerName(params.GetCocoonCodeId(), params.GetName())

	var createChainFunc func() error
	if params.Chained {
		createChainFunc = func() error {
			_, err := od.blockchain.CreateChain(name, params.Public)
			return err
		}
	}

	ledger, err := od.store.CreateLedgerThen(name, params.GetChained(), params.GetPublic(), createChainFunc)
	if err != nil {
		return nil, err
	}

	// replace hashed name to user readable name
	ledger.Name = params.GetName()

	ledgerJSON, _ := util.ToJSON(ledger)
	var protoLedger proto.Ledger
	util.FromJSON(ledgerJSON, &protoLedger)

	return &protoLedger, nil
}

// GetLedger returns a ledger
func (od *Orderer) GetLedger(ctx context.Context, params *proto.GetLedgerParams) (*proto.Ledger, error) {

	name := od.store.MakeLedgerName(params.GetCocoonCodeId(), params.GetName())
	ledger, err := od.store.GetLedger(name)
	if err != nil {
		return nil, err
	} else if ledger == nil && err == nil {
		return nil, types.ErrLedgerNotFound
	}

	// replace hashed name to user readable name
	ledger.Name = params.GetName()

	ledgerJSON, _ := util.ToJSON(ledger)
	var protoLedger proto.Ledger
	util.FromJSON(ledgerJSON, &protoLedger)

	return &protoLedger, nil
}

// Put creates a new transaction
func (od *Orderer) Put(ctx context.Context, params *proto.PutTransactionParams) (*proto.PutResult, error) {

	start := time.Now()

	// check if ledger exists
	ledgerName := od.store.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedgerName())
	ledger, err := od.store.GetLedger(ledgerName)
	if err != nil {
		return nil, err
	} else if err == nil && ledger == nil {
		return nil, types.ErrLedgerNotFound
	}

	blockID := util.Sha256(util.UUID4())
	for _, tx := range params.GetTransactions() {
		tx.Key = od.store.MakeTxKey(params.GetCocoonCodeId(), tx.Key)
		tx.BlockId = blockID
	}

	// convert []proto.Transaction to []types.Transaction
	txsAsJSON, _ := util.ToJSON(params.GetTransactions())
	var transactions []*types.Transaction
	util.FromJSON(txsAsJSON, &transactions)

	var block *proto.Block
	var createBlockFunc func() error
	if ledger.Chained {
		block = &proto.Block{}

		createBlockFunc = func() error {
			b, err := od.blockchain.CreateBlock(blockID, ledgerName, transactions)
			if b != nil {
				block.Id = b.ID
				block.ChainName = b.ChainName
				block.Hash = b.Hash
				block.Number = int64(b.Number)
				block.PrevBlockHash = b.PrevBlockHash
				block.CreatedAt = b.CreatedAt
			}
			return err
		}
	}

	err = od.store.PutThen(ledgerName, transactions, createBlockFunc)
	if err != nil {
		return nil, err
	}

	log.Debug("Put(): Time taken: ", time.Since(start))

	return &proto.PutResult{
		Added: int32(len(transactions)),
		Block: block,
	}, nil
}

// Get returns a transaction with a matching key
func (od *Orderer) Get(ctx context.Context, params *proto.GetParams) (*proto.Transaction, error) {

	start := time.Now()

	ledger, err := od.GetLedger(ctx, &proto.GetLedgerParams{
		CocoonCodeId: params.GetCocoonCodeId(),
		Name:         params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	ledgerName := od.store.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedger())
	key := od.store.MakeTxKey(params.GetCocoonCodeId(), params.GetKey())
	tx, err := od.store.Get(ledgerName, key)
	if err != nil {
		return nil, err
	} else if tx == nil && err == nil {
		return nil, types.ErrTxNotFound
	}

	if ledger.Chained {
		block, err := od.blockchain.GetBlock(ledgerName, tx.BlockID)
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf("failed to populate block to transaction")
		} else if block == nil && err == nil {
			return nil, fmt.Errorf("orphaned transaction")
		}

		tx.Block = block
		tx.BlockID = ""
	}

	txJSON, _ := util.ToJSON(tx)
	var protoTx proto.Transaction
	util.FromJSON(txJSON, &protoTx)

	log.Debug("Get(): Time taken: ", time.Since(start))

	return &protoTx, nil
}

// GetByID finds and returns a transaction with a matching id
func (od *Orderer) GetByID(ctx context.Context, params *proto.GetParams) (*proto.Transaction, error) {

	ledger, err := od.GetLedger(ctx, &proto.GetLedgerParams{
		CocoonCodeId: params.GetCocoonCodeId(),
		Name:         params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	ledgerName := od.store.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedger())
	tx, err := od.store.GetByID(ledgerName, params.GetId())
	if err != nil {
		return nil, err
	} else if tx == nil && err == nil {
		return nil, types.ErrTxNotFound
	}

	if ledger.Chained {
		block, err := od.blockchain.GetBlock(ledgerName, tx.BlockID)
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf("failed to populate block to transaction")
		} else if block == nil && err == nil {
			return nil, fmt.Errorf("orphaned transaction")
		}

		tx.Block = block
		tx.BlockID = ""
	}

	txJSON, _ := util.ToJSON(tx)
	var protoTx proto.Transaction
	util.FromJSON(txJSON, &protoTx)

	return &protoTx, nil
}

// GetBlockByID returns a block by its id and chain/ledger name
func (od *Orderer) GetBlockByID(ctx context.Context, params *proto.GetBlockParams) (*proto.Block, error) {

	name := od.store.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedger())
	ledger, err := od.store.GetLedger(name)
	if err != nil {
		return nil, err
	} else if ledger == nil && err == nil {
		return nil, types.ErrLedgerNotFound
	}

	blk, err := od.blockchain.GetBlock(name, params.GetId())
	if err != nil {
		return nil, err
	} else if blk == nil && err == nil {
		return nil, types.ErrBlockNotFound
	}

	blkJSON, _ := util.ToJSON(blk)
	var protoBlk proto.Block
	util.FromJSON(blkJSON, &protoBlk)

	return &protoBlk, nil
}

// GetRange fetches transactions between a range of keys
func (od *Orderer) GetRange(ctx context.Context, params *proto.GetRangeParams) (*proto.Transactions, error) {

	name := od.store.MakeLedgerName(params.GetCocoonCodeId(), params.GetLedger())
	ledger, err := od.store.GetLedger(name)
	if err != nil {
		return nil, err
	} else if ledger == nil && err == nil {
		return nil, types.ErrLedgerNotFound
	}

	if len(params.GetStartKey()) > 0 {
		params.StartKey = od.store.MakeTxKey(params.GetCocoonCodeId(), params.GetStartKey())
	}

	if len(params.GetEndKey()) > 0 {
		if len(params.GetStartKey()) > 0 {
			params.EndKey = od.store.MakeTxKey(params.GetCocoonCodeId(), params.GetEndKey())
		} else {
			params.EndKey = od.store.MakeTxKey(params.GetCocoonCodeId(), "%"+params.GetEndKey())
		}
	}

	txs, err := od.store.GetRange(name, params.GetStartKey(), params.GetEndKey(), params.GetInclusive(), int(params.GetLimit()), int(params.GetOffset()))
	if err != nil {
		return nil, err
	}

	txsJSON, _ := util.ToJSON(txs)
	var protoTxs []*proto.Transaction
	util.FromJSON(txsJSON, &protoTxs)

	return &proto.Transactions{
		Transactions: protoTxs,
	}, nil
}
