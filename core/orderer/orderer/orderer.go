package orderer

import (
	"fmt"
	"net"
	"os"
	"strings"

	context "golang.org/x/net/context"

	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/orderer/proto_orderer"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = config.MakeLogger("orderer", "orderer")

// SetLogLevel sets the log level of the logger
func SetLogLevel(l logging.Level) {
	logging.SetLevel(l, log.Module)
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

		if od.blockchain == nil {
			log.Error("Blockchain implementation not set")
			od.Stop(1)
			return
		}

		// establish connection to blockchain backend
		if od.blockchain != nil {
			_, err = od.blockchain.Connect(storeConStr)
			if err != nil {
				log.Info(err.Error())
				od.Stop(1)
				return
			}
		}

		// initialize the blockchain
		err = od.blockchain.Init()
		if err != nil {
			log.Info(err.Error())
			od.Stop(1)
			return
		}

		// initialize store
		if od.store == nil {
			log.Error("Store implementation not set")
			od.Stop(1)
			return
		}

		// establish connection to store backend
		_, err := od.store.Connect(storeConStr)
		if err != nil {
			log.Info(err.Error())
			od.Stop(1)
			return
		}

		if len(os.Getenv("DEV_MEM_LOCK")) != 0 {
			log.Debug("Memory based lock is in use")
		}

		sysPubLedger := types.MakeLedgerName(types.SystemCocoonID, types.GetSystemPublicLedgerName())
		sysPrivLedger := types.MakeLedgerName(types.SystemCocoonID, types.GetSystemPrivateLedgerName())
		od.store.SetBlockchainImplementation(od.blockchain)
		err = od.store.Init(sysPubLedger, sysPrivLedger)
		if err != nil {
			log.Info(err.Error())
			od.Stop(1)
			return
		}

		log.Info("Backend successfully connected")
	})

	od.server = grpc.NewServer()
	proto_orderer.RegisterOrdererServer(od.server, od)
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
	log.Infof("Setting store implementation named %s", ch.GetImplementationName())
	od.store = ch
}

// SetBlockchain sets the blockchain implementation
func (od *Orderer) SetBlockchain(b types.Blockchain) {
	log.Infof("Setting blockchain implementation named %s", b.GetImplementationName())
	od.blockchain = b
}

// CreateLedger creates a new ledger
func (od *Orderer) CreateLedger(ctx context.Context, params *proto_orderer.CreateLedgerParams) (*proto_orderer.Ledger, error) {

	internalName := types.MakeLedgerName(params.GetCocoonID(), params.GetName())
	ledger, err := od.store.CreateLedger(params.CocoonID, internalName, params.GetChained(), params.GetPublic())
	if err != nil {
		return nil, err
	}

	ledger.Name = params.GetName()
	ledger.NameInternal = internalName

	var protoLedger proto_orderer.Ledger
	cstructs.Copy(ledger, &protoLedger)

	return &protoLedger, nil
}

// GetLedger returns a ledger
func (od *Orderer) GetLedger(ctx context.Context, params *proto_orderer.GetLedgerParams) (*proto_orderer.Ledger, error) {

	internalName := types.MakeLedgerName(params.GetCocoonID(), params.GetName())
	ledger, err := od.store.GetLedger(internalName)
	if err != nil {
		return nil, err
	} else if ledger == nil && err == nil {
		return nil, types.ErrLedgerNotFound
	}

	ledger.Name = params.GetName()
	ledger.NameInternal = internalName

	var protoLedger proto_orderer.Ledger
	cstructs.Copy(ledger, &protoLedger)

	return &protoLedger, nil
}

// Put creates a new transaction
func (od *Orderer) Put(ctx context.Context, params *proto_orderer.PutTransactionParams) (*proto_orderer.PutResult, error) {

	start := time.Now()

	// check if ledger exists
	internalLedgerName := types.MakeLedgerName(params.GetCocoonID(), params.GetLedgerName())
	ledger, err := od.store.GetLedger(internalLedgerName)
	if err != nil {
		return nil, err
	} else if err == nil && ledger == nil {
		return nil, types.ErrLedgerNotFound
	}

	// copy individual tx from []proto_orderer.Transaction to []types.Transaction
	// and set transactions key and block id
	blockID := util.Sha256(util.UUID4())
	var transactions = make([]*types.Transaction, len(params.GetTransactions()))
	for i, protoTx := range params.GetTransactions() {
		transactions[i] = &types.Transaction{
			Ledger:     protoTx.Ledger,
			ID:         protoTx.Id,
			Key:        types.MakeTxKey(params.GetCocoonID(), protoTx.Key),
			Value:      protoTx.Value,
			RevisionTo: protoTx.RevisionTo,
			CreatedAt:  protoTx.CreatedAt,
		}
		if ledger.Chained {
			transactions[i].BlockID = blockID
		}
	}

	// Create sub routine for PutThen() to create a block that includes all the transactions
	// that have been successfully stored in transaction created by PutThen().
	// If an empty transaction list is returned, this means they failed to be included in
	// the native postgres transactions and as such, we simply return an error which will cause
	// the PutThen() method to rollback the native Postgres transaction
	var block *proto_orderer.Block
	var createBlockFunc func(validTransactions []*types.Transaction) error
	if ledger.Chained {
		block = &proto_orderer.Block{}
		createBlockFunc = func(validTransactions []*types.Transaction) error {

			var err error

			if len(validTransactions) == 0 {
				return fmt.Errorf("no valid transaction to add to block")
			}

			common.ReRunOnError(func() error {
				b, err := od.blockchain.CreateBlock(blockID, internalLedgerName, validTransactions)
				if b != nil {
					block.Id = b.ID
					block.ChainName = b.ChainName
					block.Hash = b.Hash
					block.Number = int64(b.Number)
					block.PrevBlockHash = b.PrevBlockHash
					block.Transactions = b.Transactions
					block.CreatedAt = b.CreatedAt
				}
				// If error is not a duplicate previous block hash error, don't re-run.
				// return nil to end the re-run routine
				if err != nil && !types.IsDuplicatePrevBlockHashError(err) {
					return nil
				}
				return err
			}, 5, time.Duration(2)*time.Second)
			return err
		}
	}

	txReceipts, err := od.store.PutThen(internalLedgerName, transactions, createBlockFunc)
	if err != nil {
		log.Errorf("failed to PUT: %s", err.Error())
		return nil, err
	}

	var protoTxReceipts = make([]*proto_orderer.TxReceipt, len(txReceipts))
	for i, r := range txReceipts {
		protoTxReceipts[i] = &proto_orderer.TxReceipt{ID: r.ID, Err: r.Err}
	}

	log.Debug("Put(): Time taken: ", time.Since(start))

	return &proto_orderer.PutResult{
		TxReceipts: protoTxReceipts,
		Block:      block,
	}, nil
}

// Get returns a transaction with a matching key
func (od *Orderer) Get(ctx context.Context, params *proto_orderer.GetParams) (*proto_orderer.Transaction, error) {

	start := time.Now()
	ledger, err := od.GetLedger(ctx, &proto_orderer.GetLedgerParams{
		CocoonID: params.GetCocoonID(),
		Name:     params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	key := types.MakeTxKey(params.GetCocoonID(), params.GetKey())
	tx, err := od.store.Get(ledger.NameInternal, key)
	if err != nil {
		return nil, err
	}
	if tx == nil && err == nil {
		return nil, types.ErrTxNotFound
	}

	tx.Key = params.GetKey()
	tx.KeyInternal = key
	tx.Ledger = params.GetLedger()
	tx.LedgerInternal = ledger.NameInternal
	if ledger.Chained {
		block, err := od.blockchain.GetBlock(ledger.NameInternal, tx.BlockID)
		if err != nil {
			return nil, fmt.Errorf("failed to populate block to transaction")
		} else if block == nil && err == nil {
			return nil, fmt.Errorf("orphaned transaction")
		}

		tx.Block = block
		tx.BlockID = ""
	}

	var protoTx proto_orderer.Transaction
	cstructs.Copy(tx, &protoTx)

	log.Debug("Get(): Time taken: ", time.Since(start))

	return &protoTx, nil
}

// GetBlockByID returns a block by its id and chain/ledger name
func (od *Orderer) GetBlockByID(ctx context.Context, params *proto_orderer.GetBlockParams) (*proto_orderer.Block, error) {

	ledger, err := od.GetLedger(ctx, &proto_orderer.GetLedgerParams{
		CocoonID: params.GetCocoonID(),
		Name:     params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	blk, err := od.blockchain.GetBlock(ledger.NameInternal, params.GetId())
	if err != nil {
		return nil, err
	} else if blk == nil && err == nil {
		return nil, types.ErrBlockNotFound
	}

	var protoBlk proto_orderer.Block
	cstructs.Copy(blk, &protoBlk)

	return &protoBlk, nil
}

// GetRange fetches transactions between a range of keys
func (od *Orderer) GetRange(ctx context.Context, params *proto_orderer.GetRangeParams) (*proto_orderer.Transactions, error) {

	ledger, err := od.GetLedger(ctx, &proto_orderer.GetLedgerParams{
		CocoonID: params.GetCocoonID(),
		Name:     params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	if len(params.GetStartKey()) > 0 {
		params.StartKey = types.MakeTxKey(params.GetCocoonID(), params.GetStartKey())
	}

	if len(params.GetEndKey()) > 0 {
		if len(params.GetStartKey()) > 0 {
			params.EndKey = types.MakeTxKey(params.GetCocoonID(), params.GetEndKey())
		} else {
			params.EndKey = types.MakeTxKey(params.GetCocoonID(), "%"+params.GetEndKey())
		}
	}

	txs, err := od.store.GetRange(ledger.NameInternal, params.GetStartKey(), params.GetEndKey(), params.GetInclusive(), int(params.GetLimit()), int(params.GetOffset()))
	if err != nil {
		return nil, err
	}

	// fetch transaction blocks and copy individual tx from []types.Transaction to []proto_orderer.Transaction
	var protoTxs = make([]*proto_orderer.Transaction, len(txs))
	for i, tx := range txs {

		if ledger.Chained {
			block, err := od.blockchain.GetBlock(ledger.NameInternal, tx.BlockID)
			if err != nil {
				return nil, fmt.Errorf("failed to populate block to transaction")
			} else if block == nil && err == nil {
				return nil, fmt.Errorf("orphaned transaction")
			}

			tx.Block = block
			tx.BlockID = ""
		}

		tx.KeyInternal = tx.Key
		tx.Key = types.GetActualKeyFromTxKey(tx.Key)
		tx.LedgerInternal = tx.Ledger
		tx.Ledger = params.GetLedger()
		var protoTx = proto_orderer.Transaction{}
		cstructs.Copy(tx, &protoTx)
		protoTxs[i] = &protoTx
	}

	return &proto_orderer.Transactions{
		Transactions: protoTxs,
	}, nil
}
