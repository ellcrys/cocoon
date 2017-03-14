package impl

import (
	"fmt"
	"strings"
	"time"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cocoon/core/types/blockchain"
	"github.com/ncodes/cocoon/core/types/store"
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
		if common.IsUniqueConstraintError(err, "chain_name") {
			return nil, fmt.Errorf("chain with matching name already exists")
		}
		return nil, err
	}

	tx.Commit()
	return newChain, nil
}

// GetChain gets a chain
func (b *PostgresBlockchain) GetChain(name string) (*blockchain.Chain, error) {

	var chain blockchain.Chain
	err := b.db.Where("name = ?", name).Find(&chain).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get chain. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &chain, nil
}

// MakeChainName returns a new chain name prefixed with a namespace
func (b *PostgresBlockchain) MakeChainName(namespace, name string) string {
	if name == blockchain.GetGlobalChainName() {
		namespace = ""
	}
	return util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
}

// MakeTxsHash takes a slice of transactions and returns a SHA256
// hash of the transactions
func MakeTxsHash(txs []*store.Transaction) string {
	var txHashes = make([]string, len(txs))
	for _, tx := range txs {
		txHashes = append(txHashes, tx.Hash)
	}
	return util.Sha256(strings.Join(txHashes, "."))
}

// VerifyTxs checks whether the hashs of a trnsactions are valid hashs
/// based on the hash algorithm defined by Transaction.MakeHash.
func VerifyTxs(txs []*store.Transaction) (*store.Transaction, bool) {
	for _, tx := range txs {
		if tx.Hash != tx.MakeHash() {
			return tx, false
		}
	}
	return nil, true
}

// CreateBlock creates a new block. It creates a chained structure by setting the new block's previous hash
// value to the has of the last block of the chain specified. The new block's hash is calculated from the hash of
// all the contained transaction hashes.
func (b *PostgresBlockchain) CreateBlock(chainName string, transactions []*store.Transaction) (*blockchain.Block, error) {

	chain, err := b.GetChain(chainName)
	if err != nil {
		return nil, err
	} else if err == nil && chain == nil {
		return nil, types.ErrChainNotFound
	}

	if len(transactions) == 0 {
		return nil, types.ErrZeroTransactions
	}

	failedTx, validTxs := VerifyTxs(transactions)
	if !validTxs {
		return nil, fmt.Errorf("Invalid block transaction; Transaction (%s) has an invalid hash", failedTx.ID)
	}

	dbTx := b.db.Begin()
	err = dbTx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	var dummyBlock = blockchain.Block{
		Hash: strings.Repeat("0", 64),
	}

	// get last block of chain
	// I included the `has_right_sibling` check to ensure competing transactions are exit/rollback when
	// when this field is updated by the first transaction to update it. This is a safeguard for when multiple blocks try to
	// use the last block as their previous block.
	var lastBlock blockchain.Block
	err = dbTx.Where("chain_name = ? AND has_right_sibling = ?", chainName, false).Last(&lastBlock).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		dbTx.Rollback()
		return nil, fmt.Errorf("failed to get last block of chain `%s`. %s", chainName, err)
	} else if err != nil && err == gorm.ErrRecordNotFound { // no last block, use dummy block
		lastBlock = dummyBlock
	}

	var curBlockCount uint
	if err = dbTx.Model(&blockchain.Block{}).Where("chain_name = ?", chainName).Count(&curBlockCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get chain `%s` block count. %s", chainName, err)
	}

	// if last lock is not a dummy block, set has right sibling to true.
	// This will effective cause all other concurrent attempt to attach a block to the last block
	// to immediately be aborted.
	if lastBlock.Hash != dummyBlock.Hash {
		lastBlock.HasRightSibling = true
		if err = dbTx.Save(&lastBlock).Error; err != nil {
			dbTx.Rollback()
			return nil, fmt.Errorf("failed to update previous block right sibling column. %s", err)
		}
	}

	txToJSONBytes, _ := util.ToJSON(transactions)
	newBlock := &blockchain.Block{
		ID:            util.Sha256(util.UUID4()),
		Number:        curBlockCount + 1,
		ChainName:     chainName,
		PrevBlockHash: lastBlock.Hash,
		Hash:          MakeTxsHash(transactions),
		Transactions:  txToJSONBytes,
		CreatedAt:     time.Now().Unix(),
	}

	if err = dbTx.Create(&newBlock).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("failed to create block. %s", err)
	}

	dbTx.Commit()
	return newBlock, nil
}

// GetBlock fetches a block by its id
func (b *PostgresBlockchain) GetBlock(id string) (*blockchain.Block, error) {

	var block blockchain.Block
	err := b.db.Where("id = ?", id).First(&block).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get block. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &block, nil
}
