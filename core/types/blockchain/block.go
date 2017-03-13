package blockchain

// Block represents a table of chains
type Block struct {
	Number       uint     `gorm:"primary_key"`
	ChainName    string   `json:"chainName"`
	PrevHash     string   `json:"prev_hash" gorm:"type:varchar(64);unique_index:idx_name_prev_hash"`
	Hash         string   `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	Transactions []string `json:"txs" gorm:"type:text[]"`
	CreatedAt    int64    `json:"createdAt"`
}
