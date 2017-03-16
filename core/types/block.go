package types

// Block represents a table of chains
type Block struct {
	PK            uint   `json:"-" gorm:"primary_key"`
	ID            string `json:"id" gorm:"type:varchar(64);unique_index:idx_name_id"`
	Number        uint   `json:"number" sql:"DEFAULT:0"`
	ChainName     string `json:"chainName" gorm:"index:idx_name_chain_name"`
	PrevBlockHash string `json:"prevBlockHash" gorm:"type:varchar(64);unique_index:idx_name_prev_block_hash"`
	Hash          string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	Transactions  []byte `json:"txs,omitempty"`
	CreatedAt     int64  `json:"createdAt" gorm:"index:idx_name_created_at"`
}
