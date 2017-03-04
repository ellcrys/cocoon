package impl

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ncodes/cocoon/core/ledgerchain/types"
)

// ErrChainExist represents an error about an existing chain
var ErrChainExist = errors.New("chain already exists")

// LedgerTableName represents the name of the table where all
// known ledger info are stored.
const LedgerTableName = "ledgers"

// TransactionTableName represents the name of the table where all
// transactions are stored.
const TransactionTableName = "transactions"

// PostgresLedgerChain defines a ledgerchain implementation
// on the postgres database. It implements the LedgerChain interface
type PostgresLedgerChain struct {
	db *gorm.DB
}

// GetBackend returns the database backend this chain
// implementation depends on.
func (ch *PostgresLedgerChain) GetBackend() string {
	return "postgres"
}

// Connect connects to a postgress server and returns a client
// or error if connection failed.
func (ch *PostgresLedgerChain) Connect(dbAddr string) (interface{}, error) {

	var err error
	ch.db, err = gorm.Open("postgres", dbAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ledgerchain backend")
	}

	ch.db.LogMode(false)

	return ch.db, nil
}

// MakeLegderHash takes a ledger and computes a hash
func (ch *PostgresLedgerChain) MakeLegderHash(ledger *types.Ledger) string {
	return util.Sha256(fmt.Sprintf("%s|%t|%d|%s", ledger.Name, ledger.Public, ledger.CreatedAt, ledger.PrevLedgerHash))
}

// MakeTxHash creates a hash of a transaction
func (ch *PostgresLedgerChain) MakeTxHash(tx *types.Transaction) string {
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
func (ch *PostgresLedgerChain) Init(globalLedgerName string) error {

	// create ledger table if not existing
	if !ch.db.HasTable(LedgerTableName) {
		if err := ch.db.CreateTable(&types.Ledger{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", LedgerTableName, err)
		}
	}

	// create transaction table if not existing
	if !ch.db.HasTable(TransactionTableName) {
		if err := ch.db.CreateTable(&types.Transaction{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", TransactionTableName, err)
		}
	}

	// Create global ledger if it does not exists
	var c int
	if err := ch.db.Model(&types.Ledger{}).Count(&c).Error; err != nil {
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

// isUniqueConstraintError checks whether an error is a postgres
// contraint error affecting a column.
func isUniqueConstraintError(err error, column string) bool {
	if m, _ := regexp.Match(`^.*unique constraint "idx_name_`+column+`"$`, []byte(err.Error())); m {
		return true
	}
	return false
}

// CreateLedger creates a new ledger.
func (ch *PostgresLedgerChain) CreateLedger(name string, public bool) (*types.Ledger, error) {

	tx := ch.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newLedger := &types.Ledger{
		Name:           name,
		Public:         public,
		NextLedgerHash: "",
		CreatedAt:      time.Now().Unix(),
	}

	var prevLedger types.Ledger
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
func (ch *PostgresLedgerChain) GetLedger(name string) (*types.Ledger, error) {

	var l types.Ledger

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
func (ch *PostgresLedgerChain) Put(txID, ledger, key, value string) (*types.Transaction, error) {

	tx := ch.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newTx := &types.Transaction{
		ID:        txID,
		Ledger:    ledger,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now().Unix(),
	}

	var prevTx types.Transaction
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
func (ch *PostgresLedgerChain) GetByID(txID string) (*types.Transaction, error) {
	var tx types.Transaction

	err := ch.db.Where("id = ?", txID).First(&tx).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to perform find op. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &tx, nil
}

// Get fetches a transaction by its ledger and key
func (ch *PostgresLedgerChain) Get(ledger, key string) (*types.Transaction, error) {
	var tx types.Transaction

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
func (ch *PostgresLedgerChain) MakeLedgerName(namespace, name string) string {
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}

// MakeTxKey creates a transaction key name for use for creating or querying a transaction.
// Accepts a namespace value and the key name.
func (ch *PostgresLedgerChain) MakeTxKey(namespace, name string) string {
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}

// Close releases any resource held
func (ch *PostgresLedgerChain) Close() error {
	if ch.db != nil {
		return ch.db.Close()
	}
	return nil
}
