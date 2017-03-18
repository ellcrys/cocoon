package types

import "github.com/ellcrys/util"
import "fmt"

// Block represents a table of chains
type Block struct {
	PK            uint   `json:"-" gorm:"primary_key"`
	ID            string `json:"id" gorm:"type:varchar(64);unique_index:idx_name_id"`
	Number        uint   `json:"number" sql:"DEFAULT:0"`
	ChainName     string `json:"chainName" gorm:"index:idx_name_chain_name"`
	PrevBlockHash string `json:"prevBlockHash" gorm:"type:varchar(64);unique_index:idx_name_prev_block_hash"`
	Hash          string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	Transactions  []byte `json:"transactions,omitempty"`
	CreatedAt     int64  `json:"createdAt" gorm:"index:idx_name_created_at"`
}

// GetTransactions returns a slice of transactions in the block.
func (b *Block) GetTransactions() ([]*Transaction, error) {
	var txs []*Transaction
	if err := util.FromJSON(b.Transactions, &txs); err != nil {
		return nil, fmt.Errorf("failed to parse transactions")
	}
	return txs, nil
}
