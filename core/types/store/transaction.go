package store

// Transaction reprents a group of transactions belonging to a ledger.
// All transaction entries must reference the hash of the immediate transaction
// sharing the same ledger name.
type Transaction struct {
	Number     uint   `json:"number" gorm:"primary_key"`
	Ledger     string `json:"ledger" gorm:"type:varchar(64)"`
	ID         string `json:"id" gorm:"type:varchar(64);unique_index:idx_name_id"`
	Key        string `json:"key" gorm:"type:varchar(64)"`
	Value      string `json:"value" gorm:"type:text"`
	Hash       string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	PrevTxHash string `json:"prevTxHash" gorm:"type:varchar(64);unique_index:idx_name_prev_tx_hash" sql:"default:null"`
	NextTxHash string `json:"nextTxHash" gorm:"type:varchar(64);unique_index:idx_name_next_tx_hash"`
	CreatedAt  int64  `json:"createdAt"`
}
