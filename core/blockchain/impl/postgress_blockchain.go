package impl

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types/blockchain"
)

// ChainTableName is the name of the table holding all chain information
var ChainTableName = "chains"

// BlockTableName is the name of the transaction block table
var BlockTableName = "blocks"

// PostgresBlockchain implements the Blockchain inyerface
type PostgresBlockchain struct {
	db *gorm.DB
}

// GetImplmentationName returns th
func (b *PostgresBlockchain) GetImplmentationName() string {
	return "postgres.blockchain"
}

// Connect connects to a postgress server and returns a client
// or error if connection failed.
func (b *PostgresBlockchain) Connect(dbAddr string) (interface{}, error) {

	var err error
	b.db, err = gorm.Open("postgres", dbAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain backend")
	}

	b.db.LogMode(false)

	return b.db, nil
}

// Init initializes the blockchain. Creates the necessary tables.
func (b *PostgresBlockchain) Init(globalChainName string) error {

	// create ledger table if not existing
	if !b.db.HasTable(ChainTableName) {
		if err := b.db.CreateTable(&blockchain.Chain{}).Error; err != nil {
			return fmt.Errorf("failed to create blockchain `%s` table. %s", ChainTableName, err)
		}
	}

	// create block table if not existing
	if !b.db.HasTable(BlockTableName) {
		if err := b.db.CreateTable(&blockchain.Block{}).Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", BlockTableName, err)
		}
	}

	// Create global vhain if it does not exists
	var c int
	if err := b.db.Model(&blockchain.Chain{}).Count(&c).Error; err != nil {
		return fmt.Errorf("failed to check whether global blockchain exists in the chain list. %s", err)
	}

	if c == 0 {
		_, err := b.CreateChain(globalChainName, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateChain creates a new chain
func (b *PostgresBlockchain) CreateChain(name string, public bool) (*blockchain.Chain, error) {

	tx := b.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newChain := &blockchain.Chain{
		Name:      name,
		Public:    public,
		CreatedAt: time.Now().Unix(),
	}

	if err := tx.Create(newChain).Error; err != nil {
		tx.Rollback()
		if common.IsUniqueConstraintError(err, "name") {
			return nil, fmt.Errorf("chain with matching name already exists")
		}
		return nil, err
	}

	tx.Commit()
	return newChain, nil
}

// MakeChainName returns a new chain name prefixed with a namespace
func (b *PostgresBlockchain) MakeChainName(namespace, name string) string {
	if name == blockchain.GetGlobalChainName() {
		namespace = ""
	}
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}
