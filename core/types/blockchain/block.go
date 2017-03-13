package blockchain

// Block represents a table of chains
type Block struct {
	ID              uint   `gorm:"primary_key"`
	Number          uint   `json:"number" sql:"DEFAULT:0"`
	ChainName       string `json:"chainName"`
	PrevBlockHash   string `json:"prevBlockHash" gorm:"type:varchar(64);unique_index:idx_name_prev_hash"`
	Hash            string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	Transactions    []byte `json:"txs"`
	HasRightSibling bool   `json:"hasRightSibling"`
	CreatedAt       int64  `json:"createdAt"`
}
