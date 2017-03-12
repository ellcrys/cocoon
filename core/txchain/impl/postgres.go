package impl

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"os"

	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ncodes/cocoon/core/types/txchain"
)

// ErrChainExist represents an error about an existing chain
var ErrChainExist = errors.New("chain already exists")

// LedgerTableName represents the name of the table where all
// known ledger info are stored.
const LedgerTableName = "ledgers"

// TransactionTableName represents the name of the table where all
// transactions are stored.
const TransactionTableName = "transactions"

// PostgresTxChain defines a txchain implementation
// on the postgres database. It implements the TxChain interface
type PostgresTxChain struct {
	db *gorm.DB
}

// GetBackend returns the database backend this chain
// implementation depends on.
func (ch *PostgresTxChain) GetBackend() string {
	return "postgres"
}

// Connect connects to a postgress server and returns a client
// or error if connection failed.
func (ch *PostgresTxChain) Connect(dbAddr string) (interface{}, error) {

	var err error
	ch.db, err = gorm.Open("postgres", dbAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to txchain backend")
	}

	ch.db.LogMode(false)

	return ch.db, nil
}

// MakeLegderHash takes a ledger and computes a hash
func (ch *PostgresTxChain) MakeLegderHash(ledger *txchain.Ledger) string {
	return util.Sha256(fmt.Sprintf("%s|%t|%d|%s", ledger.Name, ledger.Public, ledger.CreatedAt, ledger.PrevLedgerHash))
}

// MakeTxHash creates a hash of a transaction
func (ch *PostgresTxChain) MakeTxHash(tx *txchain.Transaction) string {
	return util.Sha256(fmt.Sprintf(
		"%s|%s|%s|%s|%d",
		tx.ID,
		crypto.ToBase64([]byte(tx.Key)),
		crypto.ToBase64([]byte(tx.Value)),
		tx.PrevTxHash,
		tx.CreatedAt))
}

// Init initializes the blockchain. Creates the necessary tables such as the
// the table holding records of all ledgers and global ledger entry
func (ch *PostgresTxChain) Init(globalLedgerName string) error {

	// create ledger table if not existing
	if !ch.db.HasTable(LedgerTableName) {
		if err := ch.db.CreateTable(&txchain.Ledger{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", LedgerTableName, err)
		}
	}

	// create transaction table if not existing
	if !ch.db.HasTable(TransactionTableName) {
		if err := ch.db.CreateTable(&txchain.Transaction{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", TransactionTableName, err)
		}
	}

	// Create global ledger if it does not exists
	var c int
	if err := ch.db.Model(&txchain.Ledger{}).Count(&c).Error; err != nil {
		return fmt.Errorf("failed to check whether global ledger exists in the ledger list table. %s", err)
	}

	if c == 0 {
		_, err := ch.CreateLedger(globalLedgerName, true)
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
		return fmt.Errorf("failed to connect to txchain backend")
	}

	return db.DropTable(txchain.Ledger{}, txchain.Transaction{}).Error
}

// isUniqueConstraintError checks whether an error is a postgres
// contraint error affecting a column.
func isUniqueConstraintError(err error, column string) bool {
	if m, _ := regexp.Match(`^.*unique constraint "idx_name_`+column+`"$`, []byte(err.Error())); m {
		return true
	}
	return false
}

// CreateLedger creates a new ledger.
func (ch *PostgresTxChain) CreateLedger(name string, public bool) (*txchain.Ledger, error) {

	tx := ch.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newLedger := &txchain.Ledger{
		Name:           name,
		Public:         public,
		NextLedgerHash: "",
		CreatedAt:      time.Now().Unix(),
	}

	var prevLedger txchain.Ledger
	err = tx.Where("next_ledger_hash = ?", "").Last(&prevLedger).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, err
	}

	if err != gorm.ErrRecordNotFound {
		newLedger.PrevLedgerHash = prevLedger.Hash
	}

	newLedger.Hash = ch.MakeLegderHash(newLedger)

	if err = tx.Model(&prevLedger).Update("next_ledger_hash", newLedger.Hash).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(newLedger).Error; err != nil {
		tx.Rollback()
		if isUniqueConstraintError(err, "name") {
			return nil, fmt.Errorf("ledger with matching name already exists")
		} else if isUniqueConstraintError(err, "hash") {
			return nil, fmt.Errorf("hash is being used by another ledger")
		} else if isUniqueConstraintError(err, "prev_ledger_hash") {
			return nil, fmt.Errorf("previous ledger hash already used")
		} else if isUniqueConstraintError(err, "next_ledger_hash") {
			return nil, fmt.Errorf("next ledger hash already used")
		}
		return nil, err
	}

	tx.Commit()

	return newLedger, nil
}

// GetLedger fetches a ledger meta information
func (ch *PostgresTxChain) GetLedger(name string) (*txchain.Ledger, error) {

	var l txchain.Ledger

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
func (ch *PostgresTxChain) Put(txID, ledger, key, value string) (*txchain.Transaction, error) {

	tx := ch.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newTx := &txchain.Transaction{
		ID:        txID,
		Ledger:    ledger,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now().Unix(),
	}

	var prevTx txchain.Transaction
	err = tx.Where("next_tx_hash = ? AND ledger = ?", "", ledger).Last(&prevTx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, err
	}

	if err != gorm.ErrRecordNotFound {
		newTx.PrevTxHash = prevTx.Hash
	}

	newTx.Hash = ch.MakeTxHash(newTx)

	if err = tx.Model(&prevTx).Update("next_tx_hash", newTx.Hash).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(newTx).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return newTx, nil
}

// GetByID fetches a transaction by its transaction id
func (ch *PostgresTxChain) GetByID(txID string) (*txchain.Transaction, error) {
	var tx txchain.Transaction

	err := ch.db.Where("id = ?", txID).First(&tx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to perform find op. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &tx, nil
}

// Get fetches a transaction by its ledger and key
func (ch *PostgresTxChain) Get(ledger, key string) (*txchain.Transaction, error) {
	var tx txchain.Transaction

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
func (ch *PostgresTxChain) MakeLedgerName(namespace, name string) string {
	if name == txchain.GetGlobalLedgerName() {
		namespace = ""
	}
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}

// MakeTxKey creates a transaction key name for use for creating or querying a transaction.
// Accepts a namespace value and the key name.
func (ch *PostgresTxChain) MakeTxKey(namespace, name string) string {
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}

// Close releases any resource held
func (ch *PostgresTxChain) Close() error {
	if ch.db != nil {
		return ch.db.Close()
	}
	return nil
}
