package types

import "github.com/ellcrys/util"
import "fmt"

// Block represents a table of chains
type Block struct {
	PK            uint   `json:"-" gorm:"primary_key"`
	ID            string `json:"id,omitempty" structs:"id,omitempty" mapstructure:"id,omitempty" gorm:"type:varchar(64);unique_index:idx_name_id"`
	Number        uint   `json:"number,omitempty" structs:"number,omitempty" mapstructure:"number,omitempty" sql:"DEFAULT:0"`
	ChainName     string `json:"chainName,omitempty" structs:"chainName,omitempty" mapstructure:"chainName,omitempty" gorm:"index:idx_name_chain_name"`
	PrevBlockHash string `json:"prevBlockHash,omitempty" structs:"prevBlockHash,omitempty" mapstructure:"prevBlockHash,omitempty" gorm:"type:varchar(64);unique_index:idx_name_prev_block_hash"`
	Hash          string `json:"hash,omitempty" structs:"hash,omitempty" mapstructure:"hash,omitempty" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	Transactions  []byte `json:"transactions,omitempty" structs:"transactions,omitempty" mapstructure:"transactions,omitempty"`
	CreatedAt     int64  `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt,omitempty" gorm:"index:idx_name_created_at"`
}

// GetTransactions returns a slice of transactions in the block.
func (b *Block) GetTransactions() ([]*Transaction, error) {
	var txs []*Transaction
	if err := util.FromJSON(b.Transactions, &txs); err != nil {
		return nil, fmt.Errorf("failed to parse transactions")
	}
	return txs, nil
}

// ToJSON returns the json equivalent of this object
func (b *Block) ToJSON() []byte {
	json, _ := util.ToJSON(b)
	return json
}
