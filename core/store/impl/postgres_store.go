package impl

import (
	"errors"
	"fmt"
	"time"

	"os"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types/store"
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
func (ch *PostgresStore) GetImplmentationName() string {
	return "postgres.store"
}

// Connect connects to a postgress server and returns a client
// or error if connection failed.
func (ch *PostgresStore) Connect(dbAddr string) (interface{}, error) {

	var err error
	ch.db, err = gorm.Open("postgres", dbAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to store backend")
	}

	ch.db.LogMode(false)

	return ch.db, nil
}

// MakeLegderHash takes a ledger and computes a hash
func (ch *PostgresStore) MakeLegderHash(ledger *store.Ledger) string {
	return util.Sha256(fmt.Sprintf("%s|%t|%d", ledger.Name, ledger.Public, ledger.CreatedAt))
}

// Init initializes the store. Creates the necessary tables such as the
// the table holding records of all ledgers and global ledger entry
func (ch *PostgresStore) Init(globalLedgerName string) error {

	// create ledger table if not existing
	if !ch.db.HasTable(LedgerTableName) {
		if err := ch.db.CreateTable(&store.Ledger{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", LedgerTableName, err)
		}
	}

	// create transaction table if not existing
	if !ch.db.HasTable(TransactionTableName) {
		if err := ch.db.CreateTable(&store.Transaction{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", TransactionTableName, err)
		}
	}

	// Create global ledger if it does not exists
	var c int
	if err := ch.db.Model(&store.Ledger{}).Count(&c).Error; err != nil {
		return fmt.Errorf("failed to check whether global ledger exists in the ledger list table. %s", err)
	}

	if c == 0 {
		_, err := ch.CreateLedger(globalLedgerName, true, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// Destroy removes the database structure of the chain.
// Will only work in a test environment
func Destroy(dbAddr string) error {

	if os.Getenv("APP_ENV") != "test" {
		return fmt.Errorf("Cowardly refusing to do it! Can only call Destroy() in test environment")
	}

	db, err := gorm.Open("postgres", dbAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to store backend")
	}

	return db.DropTable(store.Ledger{}, store.Transaction{}).Error
}

// CreateLedgerThen creates a new ledger and accepts an additional operation (via the thenFunc) to be
// executed before the ledger creation transaction is committed. If the thenFunc returns an error, the
// transaction is rolled back and error returned
func (ch *PostgresStore) CreateLedgerThen(name string, chained, public bool, thenFunc func() error) (*store.Ledger, error) {

	tx := ch.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newLedger := &store.Ledger{
		Name:      name,
		Public:    public,
		Chained:   chained,
		CreatedAt: time.Now().Unix(),
	}

	newLedger.Hash = ch.MakeLegderHash(newLedger)

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
func (ch *PostgresStore) CreateLedger(name string, chained, public bool) (*store.Ledger, error) {
	return ch.CreateLedgerThen(name, chained, public, nil)
}

// GetLedger fetches a ledger meta information
func (ch *PostgresStore) GetLedger(name string) (*store.Ledger, error) {

	var l store.Ledger

	err := ch.db.Where("name = ?", name).Last(&l).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get ledger. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &l, nil
}

// Put creates a new transaction associated to a ledger.
// Returns error if ledger does not exists or nil of successful.
func (ch *PostgresStore) Put(txID, ledger, key, value string) (*store.Transaction, error) {

	tx := ch.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newTx := &store.Transaction{
		ID:        txID,
		Ledger:    ledger,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now().Unix(),
	}

	newTx.Hash = newTx.MakeHash()

	if err := tx.Create(newTx).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return newTx, nil
}

// GetByID fetches a transaction by its transaction id
func (ch *PostgresStore) GetByID(txID string) (*store.Transaction, error) {
	var tx store.Transaction

	err := ch.db.Where("id = ?", txID).First(&tx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to perform find op. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &tx, nil
}

// Get fetches a transaction by its ledger and key
func (ch *PostgresStore) Get(ledger, key string) (*store.Transaction, error) {
	var tx store.Transaction

	err := ch.db.Where("key = ? AND ledger = ?", key, ledger).Last(&tx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &tx, nil
}

// MakeLedgerName creates a ledger name for use for creating or querying a ledger.
// Accepts a namespace value and the ledger name.
// If name provided is same as the GlobalLedgerName, then no namespace is required.
func (ch *PostgresStore) MakeLedgerName(namespace, name string) string {
	if name == store.GetGlobalLedgerName() {
		namespace = ""
	}
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}

// MakeTxKey creates a transaction key name for use for creating or querying a transaction.
// Accepts a namespace value and the key name.
func (ch *PostgresStore) MakeTxKey(namespace, name string) string {
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}

// Close releases any resource held
func (ch *PostgresStore) Close() error {
	if ch.db != nil {
		return ch.db.Close()
	}
	return nil
}
