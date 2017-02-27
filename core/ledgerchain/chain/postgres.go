package chain

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

// Ledger represents a group of linked transactions
type Ledger struct {
	Hash           string `json:"string"`
	PrevLedgerHash string `json:"prev_ledger_hash"`
	Name           string `json:"name"`
	CocoonCodeID   string `json:"cocoon_code_id"`
	Public         bool   `json:"public"`
	CreatedAt      int64  `json:"created_at"`
}

// PostgresChain defines a blockchain implementation
// modelled on the postgres database. It implements
// the Chain interface
type PostgresChain struct {
	db *gorm.DB
}

// GetBackend returns the database backend this chain
// implementation depends on.
func (ch *PostgresChain) GetBackend() string {
	return "postgres"
}

// Connect connects to a postgress server and returns a client
// or error if connection failed.
func (ch *PostgresChain) Connect(dbAddr string) (interface{}, error) {
	var err error
	ch.db, err = gorm.Open("postgres", "host=localhost user=ned dbname=cocoonchain sslmode=disable password=")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain backend")
	}
	return ch.db, nil
}

// MakeLegderHash takes a ledger and computes a hash
func (ch *PostgresChain) MakeLegderHash(ledger *Ledger) string {
	return util.Sha256(fmt.Sprintf("%s|%s|%t|%d", ledger.PrevLedgerHash, ledger.Name, ledger.Public, ledger.CreatedAt))
}

// Init initializes the blockchain. Creates the necessary tables such as the
// the table holding records of all ledgers.
func (ch *PostgresChain) Init() error {

	if !ch.db.HasTable("ledgers") {
		if err := ch.db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&Ledger{}).Error; err != nil {
			return fmt.Errorf("failed to create `ledgers` table. %s", err)
		}
	}

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

	return nil
}

// CreateLedger creates a new ledger.
func (ch *PostgresChain) CreateLedger(name string) error {
	return nil
}

// Close releases any resource held
func (ch *PostgresChain) Close() error {
	if ch.db != nil {
		return ch.Close()
	}
	return nil
}
