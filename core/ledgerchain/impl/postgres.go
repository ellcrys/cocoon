package impl

import (
	"errors"
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
)

// ErrChainExist represents an error about an existing chain
var ErrChainExist = errors.New("chain already exists")

// LedgerListName represents the name of the table where all
// known ledger info are stored.
const LedgerListName = "ledgers"

// GenesisLedgerHash is the first hash assigned to the first ledger list entry
const GenesisLedgerHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Ledger represents a group of linked transactions
type Ledger struct {
	Hash           string `json:"string" gorm:"type:varchar(64);unique_index"`
	PrevLedgerHash string `json:"prev_ledger_hash" gorm:"type:varchar(100);unique_index"`
	Name           string `json:"name" gorm:"type:varchar(100);unique_index"`
	CocoonCodeID   string `json:"cocoon_code_id"`
	Public         bool   `json:"public"`
	CreatedAt      int64  `json:"created_at"`
}

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
	return ch.db, nil
}

// MakeLegderHash takes a ledger and computes a hash
func (ch *PostgresLedgerChain) MakeLegderHash(ledger *Ledger) string {
	return util.Sha256(fmt.Sprintf("%s|%s|%t|%d", ledger.PrevLedgerHash, ledger.Name, ledger.Public, ledger.CreatedAt))
}

// Init initializes the blockchain. Creates the necessary tables such as the
// the table holding records of all ledgers.
func (ch *PostgresLedgerChain) Init() error {

	// create ledger list entry table if not existing
	if !ch.db.HasTable(LedgerListName) {
		if err := ch.db.CreateTable(&Ledger{}).Error; err != nil {
			return fmt.Errorf("failed to create `ledgers` table. %s", err)
		}
	}

	// Create general ledger list entry if it does not exists
	var c int
	if err := ch.db.Model(&Ledger{}).Count(&c).Error; err != nil {
		return fmt.Errorf("failed to check whether general ledger list entry exists in the ledger list table. %s", err)
	}

	if c == 0 {
		_, err := ch.CreateLedger("general", "", true)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateLedger creates a new ledger.
func (ch *PostgresLedgerChain) CreateLedger(name, cocoonCodeID string, public bool) (*Ledger, error) {

	tx := ch.db.Begin()
	newLedger := &Ledger{
		Name:         name,
		CocoonCodeID: cocoonCodeID,
		Public:       public,
		CreatedAt:    time.Now().Unix(),
	}

	var prevLedger Ledger
	err := tx.Last(&prevLedger).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, err
	}

	if err == gorm.ErrRecordNotFound {
		newLedger.PrevLedgerHash = GenesisLedgerHash
	} else {
		newLedger.PrevLedgerHash = prevLedger.Hash
	}

	newLedger.Hash = ch.MakeLegderHash(newLedger)

	if err := ch.db.Create(newLedger).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return newLedger, nil
}

// Close releases any resource held
func (ch *PostgresLedgerChain) Close() error {
	if ch.db != nil {
		return ch.Close()
	}
	return nil
}
