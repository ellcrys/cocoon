package impl

import (
	"fmt"
	"strings"
	"time"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
)

// ChainTableName is the name of the table holding all chain information
var ChainTableName = "chains"

// BlockTableName is the name of the transaction block table
var BlockTableName = "blocks"

// PostgresBlockchain implements the Blockchain interface
type PostgresBlockchain struct {
	db *gorm.DB
}

// GetImplementationName returns th
func (b *PostgresBlockchain) GetImplementationName() string {
	return "postgres.blockchain"
}

// Connect connects to a postgres server and returns a client
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

// Init creates the chain and block tables
func (b *PostgresBlockchain) Init() error {

	// create ledger table if not existing
	if !b.db.HasTable(ChainTableName) {
		if err := b.db.CreateTable(&types.Chain{}).Error; err != nil {
			return fmt.Errorf("failed to create blockchain `%s` table. %s", ChainTableName, err)
		}
	}

	// create block table if not existing
	if !b.db.HasTable(BlockTableName) {
		if err := b.db.CreateTable(&types.Block{}).
			AddIndex("idx_name_chain_name_id", "chain_name", "id").Error; err != nil {
			return fmt.Errorf("failed to create `%s` table. %s", BlockTableName, err)
		}
	}

	return nil
}

// CreateChain creates a new chain
func (b *PostgresBlockchain) CreateChain(name string, public bool) (*types.Chain, error) {

	tx := b.db.Begin()

	err := tx.Exec(`SET TRANSACTION isolation level repeatable read`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level. %s", err)
	}

	newChain := &types.Chain{
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
func (b *PostgresBlockchain) GetChain(name string) (*types.Chain, error) {

	var chain types.Chain
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
	return fmt.Sprintf("%s;%s", namespace, name)
}

// MakeTxsHash takes a slice of transactions and returns a SHA256
// hash of the transactions
func MakeTxsHash(txs []*types.Transaction) string {
	var txHashes = make([]string, len(txs))
	for _, tx := range txs {
		txHashes = append(txHashes, tx.Hash)
	}
	return util.Sha256(strings.Join(txHashes, ";"))
}

// VerifyTxs checks whether the hash of a transactions are valid hashes
// based on the hash algorithm defined by Transaction.MakeHash.
func VerifyTxs(txs []*types.Transaction) (*types.Transaction, bool) {
	for _, tx := range txs {
		if tx.Hash != tx.MakeHash() {
			return tx, false
		}
	}
	return nil, true
}

// MakeGenesisBlockHash creates the standard hash for the first block of a chain
// which is a sha256 hash of a concatenated chain name and 64 chars of `0`.
func MakeGenesisBlockHash(chainName string) string {
	return util.Sha256(fmt.Sprintf("%s;%s", chainName, strings.Repeat("0", 64)))
}

// CreateBlock creates a new block. It creates a chained structure by setting the new block's previous hash
// value to the hash of the last block of the chain specified. The new block's hash is calculated from the hash of
// all the contained transaction hashes.
func (b *PostgresBlockchain) CreateBlock(id, chainName string, transactions []*types.Transaction) (*types.Block, error) {

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
	var dummyBlock = types.Block{
		Hash: MakeGenesisBlockHash(chainName),
	}

	// get last block of chain
	var lastBlock types.Block
	err = dbTx.Where("chain_name = ?", chainName).Last(&lastBlock).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		dbTx.Rollback()
		return nil, fmt.Errorf("failed to get last block of chain `%s`. %s", chainName, err)
	} else if err != nil && err == gorm.ErrRecordNotFound { // no last block, use dummy block
		lastBlock = dummyBlock
	}

	var curBlockCount uint
	if err = dbTx.Model(&types.Block{}).Where("chain_name = ?", chainName).Count(&curBlockCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get chain `%s` block count. %s", chainName, err)
	}

	txToJSONBytes, _ := util.ToJSON(transactions)
	newBlock := &types.Block{
		ID:            id,
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

// GetBlock fetches a block by its chain name and id
func (b *PostgresBlockchain) GetBlock(chainName, id string) (*types.Block, error) {

	var block types.Block
	err := b.db.Where("chain_name = ? AND id = ?", chainName, id).First(&block).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get block. %s", err)
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &block, nil
}
