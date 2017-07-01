package impl

import (
	"errors"
	"fmt"
	"time"

	"os"

	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/types"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/kr/pretty"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("postgres.store")

// ErrChainExist represents an error about an existing chain
var ErrChainExist = errors.New("chain already exists")

// LedgerTableName represents the name of the table where all
// known ledger info are stored.
const LedgerTableName = "ledgers"

// TransactionTableName represents the name of the table where all
// transactions are stored.
const TransactionTableName = "transactions"

// PostgresStore defines a store implementation
// on the postgres database. It implements the Store interface
type PostgresStore struct {
	db         *gorm.DB
	blockchain types.Blockchain
	locker     types.Lock
}

// SetBlockchainImplementation sets sets a reference of the blockchain implementation
func (s *PostgresStore) SetBlockchainImplementation(b types.Blockchain) {
	s.blockchain = b
}

// GetImplementationName returns the name of this store implementation
func (s *PostgresStore) GetImplementationName() string {
	return "postgres.store"
}

// Connect connects to a postgres server and returns a client
// or error if connection failed.
func (s *PostgresStore) Connect(dbAddr string) (interface{}, error) {
	var err error
	s.db, err = gorm.Open("postgres", dbAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to store backend. %s", err)
	}

	s.db.LogMode(false)

	return s.db, nil
}

// Init initializes the types. Creates the necessary tables such as the
// the tables and public and private system ledgers
func (s *PostgresStore) Init(systemPublicLedgerName, systemPrivateLedgerName string) error {

	// create ledger table if not existing
	if !s.db.HasTable(LedgerTableName) {
		if err := s.db.CreateTable(&types.Ledger{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", LedgerTableName, err)
		}
	}

	// create transaction table if not existing
	if !s.db.HasTable(TransactionTableName) {
		if err := s.db.CreateTable(&types.Transaction{}).
			AddIndex("idx_name_ledger_key_created_at", "ledger", "key", "created_at").Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", TransactionTableName, err)
		}
	}

	// create system ledgers
	var systemLedgers = [][]interface{}{
		[]interface{}{systemPublicLedgerName, true},   // public
		[]interface{}{systemPrivateLedgerName, false}, // private
	}

	for _, ledger := range systemLedgers {
		var c int
		if err := s.db.Model(&types.Ledger{}).Where("name = ?", ledger[0].(string)).Count(&c).Error; err != nil {
			return fmt.Errorf("failed to check existence of ledger named %s: %s", ledger[0].(string), err)
		}
		if c == 0 {
			_, err := s.CreateLedger(types.SystemCocoonID, ledger[0].(string), true, ledger[1].(bool))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Destroy removes the database tables.
// Will only work in a test environment (Test Only!!!)
func Destroy(dbAddr string) error {

	if os.Getenv("ENV") != "test" {
		return fmt.Errorf("Cowardly refusing to do it! Can only call Destroy() in test environment")
	}

	db, err := gorm.Open("postgres", dbAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to store backend. %s", err)
	}

	return db.DropTable(types.Ledger{}, types.Transaction{}).Error
}

// Clear the database tables.
// Will only work in a test environment (Test Only!!!)
func Clear(dbAddr string) error {

	if os.Getenv("ENV") != "test" {
		return fmt.Errorf("Cowardly refusing to do it! Can only call Destroy() in test environment")
	}

	db, err := gorm.Open("postgres", dbAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to store backend")
	}

	err = db.Delete(types.Transaction{}).Error
	if err != nil {
		return err
	}

	err = db.Delete(types.Ledger{}).Error
	if err != nil {
		return err
	}

	return nil
}

// CreateLedgerThen creates a new ledger and accepts an additional operation (via the thenFunc) to be
// executed before the ledger creation transaction is committed. If the thenFunc returns an error, the
// transaction is rolled back and error returned
func (s *PostgresStore) CreateLedgerThen(cocoonID, name string, chained, public bool, thenFunc func() error) (*types.Ledger, error) {

	tx := s.db.Begin()

	newLedger := &types.Ledger{
		Name:      name,
		Public:    public,
		Chained:   chained,
		CocoonID:  cocoonID,
		CreatedAt: time.Now().Unix(),
	}

	if err := tx.Create(newLedger).Error; err != nil {
		tx.Rollback()
		if common.IsUniqueConstraintError(err, "name") {
			return nil, fmt.Errorf("ledger with matching name already exists")
		} else if common.IsUniqueConstraintError(err, "hash") {
			return nil, fmt.Errorf("hash is being used by another ledger")
		}
		return nil, err
	}

	// create chain
	if chained {
		_, err := s.blockchain.CreateChain(name, public)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// run the companion functions and Rollback
	// the transaction if error was returned
	if thenFunc != nil {
		if err := thenFunc(); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return newLedger, nil
}

// CreateLedger creates a new ledger.
func (s *PostgresStore) CreateLedger(cocoonID, name string, chained, public bool) (*types.Ledger, error) {
	return s.CreateLedgerThen(cocoonID, name, chained, public, nil)
}

// GetLedger fetches a ledger meta information
func (s *PostgresStore) GetLedger(name string) (*types.Ledger, error) {

	var l types.Ledger

	err := s.db.Where("name = ?", name).Last(&l).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get ledger. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &l, nil
}

// makeTxLockKey constructs a lock key using a transaction key and ledger name
func makeTxLockKey(ledgerName, key string) string {
	return fmt.Sprintf("tx/key/%s/%s", ledgerName, key)
}

// PutThen adds transactions to the store and returns a list of transaction receipts.
// Any transaction that failed to be created will result in an error receipt being created
// and returned along with success receipts of they successfully added transactions.
// However, all transactions will be rolled back if the `thenFunc` returns error. Only the
// transaction that are successfully added will be passed to the thenFunc.
// Future work may allow the caller to determine the behaviour via an additional parameter.
func (s *PostgresStore) PutThen(ledgerName string, txs []*types.Transaction, thenFunc func(validTxss []*types.Transaction) error) ([]*types.TxReceipt, error) {

	var err error
	var validTxs []*types.Transaction
	txReceipts := []*types.TxReceipt{}

	dbTx := s.db.Begin()

	// create transactions and add transaction receipts for
	// successfully stored transactions
	for _, tx := range txs {

		// acquire lock on the transaction via its key
		lock, err := common.NewLock(makeTxLockKey(ledgerName, tx.Key))
		if err != nil {
			dbTx.Rollback()
			return nil, err
		}
		if err := lock.Acquire(); err != nil {
			if err == types.ErrLockAlreadyAcquired {
				txReceipts = append(txReceipts, &types.TxReceipt{ID: tx.ID, Err: "failed to acquire lock. object has been locked by another process"})
			} else {
				txReceipts = append(txReceipts, &types.TxReceipt{ID: tx.ID, Err: err.Error()})
			}
			continue
		}

		// For transactions requiring explit pessimistic locking, ensure the current transaction is not stale
		// when checked with the already processed transactions. If this is not done, before we call the tx.Create
		// the entire transaction be considered failed when we call tx.Commit
		isStale := false
		for _, vTx := range validTxs {
			if len(tx.RevisionTo) > 0 && tx.RevisionTo == vTx.RevisionTo && tx.KeyInternal == vTx.KeyInternal {
				isStale = true
				break
			}
		}

		if isStale {
			txReceipts = append(txReceipts, &types.TxReceipt{ID: tx.ID, Err: "stale object"})
			lock.Release()
			continue
		}

		tx.Hash = tx.MakeHash()
		tx.Ledger = ledgerName
		txReceipt := &types.TxReceipt{ID: tx.ID}
		err = dbTx.Create(tx).Error
		if err != nil {
			log.Errorf("Failed to create transaction (%s): %s", tx.ID, err)
			txReceipt.Err = err.Error()
			if common.CompareErr(err, fmt.Errorf(`pq: duplicate key value violates unique constraint "idx_name_revision_to"`)) == 0 {
				txReceipt.Err = "stale object"
			}
		} else {
			validTxs = append(validTxs, tx)
		}

		if err = lock.Release(); err != nil {
			fmt.Println("Error releasing lock: ", err)
		}

		txReceipts = append(txReceipts, txReceipt)
	}

	// run the companion functions. Rollback
	// the transactions only if error was returned
	if thenFunc != nil {
		if err = thenFunc(validTxs); err != nil {
			return txReceipts, err
		}
	}

	if err := dbTx.Debug().Commit().Error; err != nil {
		pretty.Println(dbTx.GetErrors())
		return nil, err
	}

	return txReceipts, nil
}

// Put creates one or more transactions associated to a ledger.
// Returns a list of transaction receipts and a general error.
func (s *PostgresStore) Put(ledgerName string, txs []*types.Transaction) ([]*types.TxReceipt, error) {
	return s.PutThen(ledgerName, txs, nil)
}

// Get fetches a transaction by its ledger and key
func (s *PostgresStore) Get(ledger, key string) (*types.Transaction, error) {
	var tx types.Transaction

	// acquire lock on the transaction via its key
	lock, err := common.NewLock(makeTxLockKey(ledger, key))
	if err != nil {
		return nil, err
	}
	if err := lock.Acquire(); err != nil {
		if err == types.ErrLockAlreadyAcquired {
			return nil, fmt.Errorf("failed to acquire lock. object has been locked by another process")
		}
		return nil, err
	}

	defer lock.Release()

	err = s.db.Where("ledger = ? AND key = ?", ledger, key).Last(&tx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &tx, nil
}

// GetRange fetches transactions with keys included in a specified range.
// No lock is acquired in this operation.
func (s *PostgresStore) GetRange(ledger, startKey, endKey string, inclusive bool, limit, offset int) ([]*types.Transaction, error) {

	var err error
	var txs []*types.Transaction
	var db = s.db.Debug()

	sql := `SELECT DISTINCT ON (key) * FROM "transactions"  WHERE `
	args := []interface{}{}

	if len(startKey) > 0 && len(endKey) > 0 {
		if !inclusive {
			sql += `ledger = ? AND (key >= ? AND key < ?)`
			args = append(args, ledger, startKey, endKey)
		} else {
			sql += `ledger = ? AND (key >= ? OR key <= ?)`
			args = append(args, ledger, startKey+"%", endKey+"%")
		}
	} else if len(startKey) > 0 && len(endKey) == 0 {
		sql += `ledger = ? AND key like ?`
		args = append(args, ledger, startKey+"%")
	} else {
		// setting endKey only is a little tricky as the call code may construct
		// through a secondary process or rule, so add the '%' operator will most likely
		// result in wrong query. So we just let the external decide where to put it.
		sql += `ledger = ? AND key like ?`
		args = append(args, ledger, endKey)
	}

	args = append(args, limit, offset)
	err = db.Raw(sql+` ORDER BY key, created_at desc LIMIT ? OFFSET ?`, args...).Scan(&txs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get transactions. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return txs, nil
}

// Close releases any resource held
func (s *PostgresStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
