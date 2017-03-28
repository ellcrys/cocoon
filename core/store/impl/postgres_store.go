package impl

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"os"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
)

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
	db *gorm.DB
}

// GetImplmentationName returns the name of this store implementation
func (s *PostgresStore) GetImplmentationName() string {
	return "postgres.store"
}

// Connect connects to a postgress server and returns a client
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

// MakeLegderHash takes a ledger and computes a hash
func (s *PostgresStore) MakeLegderHash(ledger *types.Ledger) string {
	return util.Sha256(fmt.Sprintf("%s|%t|%d", ledger.Name, ledger.Public, ledger.CreatedAt))
}

// Init initializes the types. Creates the necessary tables such as the
// the table holding records of all ledgers and global ledger entry
func (s *PostgresStore) Init(globalLedgerName string) error {

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

	// Create global ledger if it does not exists
	var c int
	if err := s.db.Model(&types.Ledger{}).Count(&c).Error; err != nil {
		return fmt.Errorf("failed to check whether global ledger exists in the ledger list table. %s", err)
	}

	if c == 0 {
		_, err := s.CreateLedger(globalLedgerName, true, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// Destroy removes the database tables.
// Will only work in a test environment (Test Only!!!)
func Destroy(dbAddr string) error {

	if os.Getenv("APP_ENV") != "test" {
		return fmt.Errorf("Cowardly refusing to do it! Can only call Destroy() in test environment")
	}

	db, err := gorm.Open("postgres", dbAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to store backend")
	}

	return db.DropTable(types.Ledger{}, types.Transaction{}).Error
}

// Clear truncatest the database tables.
// Will only work in a test environment (Test Only!!!)
func Clear(dbAddr string) error {

	if os.Getenv("APP_ENV") != "test" {
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
func (s *PostgresStore) CreateLedgerThen(name string, chained, public bool, thenFunc func() error) (*types.Ledger, error) {

	tx := s.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newLedger := &types.Ledger{
		Name:      name,
		Public:    public,
		Chained:   chained,
		CreatedAt: time.Now().Unix(),
	}

	newLedger.Hash = s.MakeLegderHash(newLedger)

	if err := tx.Create(newLedger).Error; err != nil {
		tx.Rollback()
		if common.IsUniqueConstraintError(err, "name") {
			return nil, fmt.Errorf("ledger with matching name already exists")
		} else if common.IsUniqueConstraintError(err, "hash") {
			return nil, fmt.Errorf("hash is being used by another ledger")
		}
		return nil, err
	}

	// run the companion functions and Rollback
	// the transaction if error was returned
	if thenFunc != nil {
		if err = thenFunc(); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return newLedger, nil
}

// CreateLedger creates a new ledger.
func (s *PostgresStore) CreateLedger(name string, chained, public bool) (*types.Ledger, error) {
	return s.CreateLedgerThen(name, chained, public, nil)
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

// PutThen creates one or more transactions associated to a ledger.
// Returns error if any of the transaction failed or nil if
// all transactions were successful added. It accepts an additional operation (via the thenFunc) to be
// executed before the transactions are committed. If the thenFunc returns an error, the
// transaction is rolled back and error is returned
func (s *PostgresStore) PutThen(ledgerName string, txs []*types.Transaction, thenFunc func() error) error {

	dbTx := s.db.Begin()
	err := dbTx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	for _, tx := range txs {
		tx.Hash = tx.MakeHash()
		tx.Ledger = ledgerName
		if err := dbTx.Create(tx).Error; err != nil {
			dbTx.Rollback()
			return err
		}
	}

	// run the companion functions and Rollback
	// the transaction if error was returned
	if thenFunc != nil {
		if err = thenFunc(); err != nil {
			dbTx.Rollback()
			return err
		}
	}

	dbTx.Commit()
	return nil
}

// Put creates one or more transactions associated to a ledger.
// Returns error if ledger does not exists, if any of the transaction failed or nil if
// all transactions were successful added.
func (s *PostgresStore) Put(ledgerName string, txs []*types.Transaction) error {
	return s.PutThen(ledgerName, txs, nil)
}

// GetByID fetches a transaction by its transaction id
func (s *PostgresStore) GetByID(ledgerName, txID string) (*types.Transaction, error) {
	var tx types.Transaction

	err := s.db.Where("ledger = ? AND  id = ?", ledgerName, txID).First(&tx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to perform find op. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &tx, nil
}

// Get fetches a transaction by its ledger and key
func (s *PostgresStore) Get(ledger, key string) (*types.Transaction, error) {
	var tx types.Transaction

	err := s.db.Where("ledger = ? AND key = ?", ledger, key).Last(&tx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &tx, nil
}

// GetRange detches transactions with keys included in a specified range.
func (s *PostgresStore) GetRange(ledger, startKey, endKey string, inclusive bool, limit, offset int) ([]*types.Transaction, error) {

	var err error
	var txs []*types.Transaction
	var q *gorm.DB

	if len(startKey) > 0 && len(endKey) > 0 {
		if !inclusive {
			q = s.db.Where("ledger = ? AND (key >= ? AND key < ?)", ledger, startKey, endKey)
		} else {
			q = s.db.Where("ledger = ? AND (key >= ? OR key <= ?)", ledger, startKey+"%", endKey+"%")
		}
	} else if len(startKey) > 0 && len(endKey) == 0 {
		q = s.db.Where("ledger = ? AND key like ?", ledger, startKey+"%")
	} else {
		// setting endKey only is a little tricky as the call code may construct
		// through a secondary process or rule, so add the '%' operator will most likely
		// result in wrong query. So we just let the external decide where to put it.
		q = s.db.Where("ledger = ? AND key like ?", ledger, endKey)
	}

	err = q.
		Limit(limit).
		Offset(offset).
		Select("DISTINCT ON (key) *").
		Order("key").
		Order("created_at desc").
		Find(&txs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get transactions. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return txs, nil
}

// MakeLedgerName creates a ledger name for use for creating or querying a ledger.
// Accepts a namespace value and the ledger name.
// If name provided is same as the GlobalLedgerName, then no namespace is required.
func (s *PostgresStore) MakeLedgerName(namespace, name string) string {
	if name == types.GetGlobalLedgerName() {
		namespace = ""
	}
	return fmt.Sprintf("%s.%s", namespace, name)
}

// MakeTxKey creates a transaction key name for use for creating or querying a transaction.
// Accepts a namespace value and the key name.
func (s *PostgresStore) MakeTxKey(namespace, name string) string {
	return fmt.Sprintf("%s.%s", namespace, name)
}

// GetActualKeyFromTxKey returns the real key name from a transaction key
func (s *PostgresStore) GetActualKeyFromTxKey(key string) string {
	return strings.Split(key, ".")[1]
}

// Close releases any resource held
func (s *PostgresStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
