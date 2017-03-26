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
	"github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("orderer")

// SetLogLevel sets the log level of the logger
func SetLogLevel(l logging.Level) {
	logging.SetLevel(l, log.Module)
}

// DiscoverOrderers fetches a list of orderer service addresses
// via consul service discovery API. For development purpose,
// If DEV_ORDERER_ADDR is set, it will fetch the orderer
// address from the env variable.
func DiscoverOrderers() ([]string, error) {

	if len(os.Getenv("DEV_ORDERER_ADDR")) > 0 {
		return []string{os.Getenv("DEV_ORDERER_ADDR")}, nil
	}

	ds := scheduler.NomadServiceDiscovery{
		ConsulAddr: util.Env("CONSUL_ADDR", "localhost:8500"),
		Protocol:   "http",
	}

	_orderers, err := ds.GetByID("orderers", nil)
	if err != nil {
		return []string{}, nil
	}

	var orderers []string
	for _, orderer := range _orderers {
		orderers = append(orderers, fmt.Sprintf("%s:%d", orderer.IP, int(orderer.Port)))
	}

	return orderers, nil
}

// DialOrderer returns a connection to a orderer from a list of addresses. It randomly
// picks an orderer address from the list for orderers.
func DialOrderer(ordererAddrs []string) (*grpc.ClientConn, error) {
	var ordererAddr string

	if len(ordererAddrs) == 0 {
		return nil, fmt.Errorf("no known orderer address")
	} else if len(ordererAddrs) == 1 {
		ordererAddr = ordererAddrs[0]
	} else {
		ordererAddr = ordererAddrs[util.RandNum(0, len(ordererAddrs))]
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

	internalName := od.store.MakeLedgerName(params.GetCocoonID(), params.GetName())

	var createChainFunc func() error
	if params.Chained {
		createChainFunc = func() error {
			_, err := od.blockchain.CreateChain(internalName, params.Public)
			return err
		}
	}

	ledger, err := od.store.CreateLedgerThen(internalName, params.GetChained(), params.GetPublic(), createChainFunc)
	if err != nil {
		return nil, err
	}

	ledger.Name = params.GetName()
	ledger.NameInternal = internalName

	var protoLedger proto.Ledger
	cstructs.Copy(ledger, &protoLedger)

	return &protoLedger, nil
}

// GetLedger returns a ledger
func (od *Orderer) GetLedger(ctx context.Context, params *proto.GetLedgerParams) (*proto.Ledger, error) {

	internalName := od.store.MakeLedgerName(params.GetCocoonID(), params.GetName())
	ledger, err := od.store.GetLedger(internalName)
	if err != nil {
		return nil, err
	} else if ledger == nil && err == nil {
		return nil, types.ErrLedgerNotFound
	}

	ledger.Name = params.GetName()
	ledger.NameInternal = internalName

	var protoLedger proto.Ledger
	cstructs.Copy(ledger, &protoLedger)

	return &protoLedger, nil
}

// Put creates a new transaction
func (od *Orderer) Put(ctx context.Context, params *proto.PutTransactionParams) (*proto.PutResult, error) {

	start := time.Now()

	// check if ledger exists
	internalLedgerName := od.store.MakeLedgerName(params.GetCocoonID(), params.GetLedgerName())
	ledger, err := od.store.GetLedger(internalLedgerName)
	if err != nil {
		return nil, err
	} else if err == nil && ledger == nil {
		return nil, types.ErrLedgerNotFound
	}

	// copy individual tx from []proto.Transaction to []types.Transaction
	// and set transactions key and block id
	blockID := util.Sha256(util.UUID4())
	var transactions = make([]*types.Transaction, len(params.GetTransactions()))
	for i, protoTx := range params.GetTransactions() {
		var tx = types.Transaction{}
		cstructs.Copy(protoTx, &tx)
		tx.Key = od.store.MakeTxKey(params.GetCocoonID(), tx.Key)
		transactions[i] = &tx
		if ledger.Chained {
			tx.BlockID = blockID
		}
	}

	var block *proto.Block
	var createBlockFunc func() error
	if ledger.Chained {
		block = &proto.Block{}

		createBlockFunc = func() error {
			var err error
			retryDelay := time.Duration(2) * time.Second
			common.ReRunOnError(func() error {
				b, _err := od.blockchain.CreateBlock(blockID, internalLedgerName, transactions)
				if b != nil {
					block.Id = b.ID
					block.ChainName = b.ChainName
					block.Hash = b.Hash
					block.Number = int64(b.Number)
					block.PrevBlockHash = b.PrevBlockHash
					block.Transactions = b.Transactions
					block.CreatedAt = b.CreatedAt
				}

				err = _err

				// If error is not a duplicate previous block hash error, don't re-run.
				// return nil to end the re-run routine
				if _err != nil && !types.IsDuplicatePrevBlockHashError(_err) {
					return nil
				}

				return _err
			}, 5, &retryDelay)
			return err
		}
	}

	err = od.store.PutThen(internalLedgerName, transactions, createBlockFunc)
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
		CocoonID: params.GetCocoonID(),
		Name:     params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	key := od.store.MakeTxKey(params.GetCocoonID(), params.GetKey())
	tx, err := od.store.Get(ledger.NameInternal, key)
	if err != nil {
		return nil, err
	} else if tx == nil && err == nil {
		return nil, types.ErrTxNotFound
	}

	tx.Key = params.GetKey()
	tx.KeyInternal = key
	tx.Ledger = params.GetLedger()
	tx.LedgerInternal = ledger.NameInternal

	if ledger.Chained {
		block, err := od.blockchain.GetBlock(ledger.NameInternal, tx.BlockID)
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf("failed to populate block to transaction")
		} else if block == nil && err == nil {
			return nil, fmt.Errorf("orphaned transaction")
		}

		tx.Block = block
		tx.BlockID = ""
	}

	var protoTx proto.Transaction
	cstructs.Copy(tx, &protoTx)

	log.Debug("Get(): Time taken: ", time.Since(start))

	return &protoTx, nil
}

// GetByID finds and returns a transaction with a matching id
func (od *Orderer) GetByID(ctx context.Context, params *proto.GetParams) (*proto.Transaction, error) {

	ledger, err := od.GetLedger(ctx, &proto.GetLedgerParams{
		CocoonID: params.GetCocoonID(),
		Name:     params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	tx, err := od.store.GetByID(ledger.NameInternal, params.GetId())
	if err != nil {
		return nil, err
	} else if tx == nil && err == nil {
		return nil, types.ErrTxNotFound
	}

	tx.KeyInternal = tx.Key
	tx.Key = od.store.GetActualKeyFromTxKey(tx.Key)
	tx.Ledger = params.GetLedger()
	tx.LedgerInternal = ledger.NameInternal

	if ledger.Chained {
		block, err := od.blockchain.GetBlock(ledger.NameInternal, tx.BlockID)
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf("failed to populate block to transaction")
		} else if block == nil && err == nil {
			return nil, fmt.Errorf("orphaned transaction")
		}

		tx.Block = block
		tx.BlockID = ""
	}

	var protoTx proto.Transaction
	cstructs.Copy(tx, &protoTx)

	return &protoTx, nil
}

// GetBlockByID returns a block by its id and chain/ledger name
func (od *Orderer) GetBlockByID(ctx context.Context, params *proto.GetBlockParams) (*proto.Block, error) {

	ledger, err := od.GetLedger(ctx, &proto.GetLedgerParams{
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

	var protoBlk proto.Block
	cstructs.Copy(blk, &protoBlk)

	return &protoBlk, nil
}

// GetRange fetches transactions between a range of keys
func (od *Orderer) GetRange(ctx context.Context, params *proto.GetRangeParams) (*proto.Transactions, error) {

	ledger, err := od.GetLedger(ctx, &proto.GetLedgerParams{
		CocoonID: params.GetCocoonID(),
		Name:     params.GetLedger(),
	})
	if err != nil {
		return nil, err
	}

	if len(params.GetStartKey()) > 0 {
		params.StartKey = od.store.MakeTxKey(params.GetCocoonID(), params.GetStartKey())
	}

	if len(params.GetEndKey()) > 0 {
		if len(params.GetStartKey()) > 0 {
			params.EndKey = od.store.MakeTxKey(params.GetCocoonID(), params.GetEndKey())
		} else {
			params.EndKey = od.store.MakeTxKey(params.GetCocoonID(), "%"+params.GetEndKey())
		}
	}

	txs, err := od.store.GetRange(ledger.NameInternal, params.GetStartKey(), params.GetEndKey(), params.GetInclusive(), int(params.GetLimit()), int(params.GetOffset()))
	if err != nil {
		return nil, err
	}

	// copy individual tx from []types.Transaction to []proto.Transaction
	var protoTxs = make([]*proto.Transaction, len(txs))
	for i, tx := range txs {
		tx.KeyInternal = tx.Key
		tx.Key = od.store.GetActualKeyFromTxKey(tx.Key)
		tx.LedgerInternal = tx.Ledger
		tx.Ledger = params.GetLedger()
		var protoTx = proto.Transaction{}
		cstructs.Copy(tx, &protoTx)
		protoTxs[i] = &protoTx
	}

	return &proto.Transactions{
		Transactions: protoTxs,
	}, nil
}
