package types

// Transaction reprents a group of transactions belonging to a ledger.
// All transaction entries must reference the hash of the immediate transaction
// sharing the same ledger name.
type Transaction struct {
	Number     uint   `gorm:"primary_key"`
	ID         string `json:"id" gorm:"type:varchar(64);unique_index"`
	Key        string `json:"key" gorm:"type:varchar(64)"`
	Value      string `json:"key" gorm:"type:text"`
	Hash       string `json:"hash" gorm:"type:varchar(64);unique_index"`
	PrevTxHash string `json:"prev_tx_hash" gorm:"type:varchar(64);unique_index"`
	NextTxHash string `json:"next_tx_hash" gorm:"type:varchar(64);unique_index"`
	CreatedAt  int64  `json:"created_at"`
}
