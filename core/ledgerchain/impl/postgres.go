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

// LedgerEntryName represents the name of the table where all
// known ledger info are stored.
const LegderEntryName = "ledgers"

// Ledger represents a group of linked transactions
type Ledger struct {
	Hash           string `json:"string"`
	PrevLedgerHash string `json:"prev_ledger_hash"`
	Name           string `json:"name"`
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

	// create ledger entry table if not existing
	if !ch.db.HasTable(LegderEntryName) {
		if err := ch.db.CreateTable(&Ledger{}).Error; err != nil {
			return fmt.Errorf("failed to create `ledgers` table. %s", err)
		}
	}

	// Create general ledger entry if it does not exists
	var c int
	if err := ch.db.Model(&Ledger{}).Count(&c).Error; err != nil {
		return fmt.Errorf("failed to check whether general ledger entry exists in the ledger entry table. %s", err)
	}

	if c == 0 {
		globalLedger := &Ledger{
			PrevLedgerHash: "genesis",
			Name:           "general",
			Public:         true,
			CreatedAt:      time.Now().Unix(),
		}

		globalLedger.Hash = ch.MakeLegderHash(globalLedger)
		if err := ch.db.Create(globalLedger).Error; err != nil {
			return fmt.Errorf("failed to create global ledger entry in ledgers table. %s", err)
		}
	}

	return nil
}

// CreateLedger creates a new ledger.
func (ch *PostgresLedgerChain) CreateLedger(name string) error {
	return nil
}

// Close releases any resource held
func (ch *PostgresLedgerChain) Close() error {
	if ch.db != nil {
		return ch.Close()
	}
	return nil
}
