package types

// Transaction reprents a group of transactions belonging to a ledger.
// All transaction entries must reference the hash of the immediate transaction
// sharing the same ledger name.
type Transaction struct {
	Number     uint   `gorm:"primary_key"`
	ID         string `gorm:"type:varchar(64);unique_index:idx_name_id"`
	Key        string `gorm:"type:varchar(64)"`
	Value      string `gorm:"type:text"`
	Hash       string `gorm:"type:varchar(64);unique_index:idx_name_hash"`
	PrevTxHash string `gorm:"type:varchar(64);unique_index:idx_name_prev_tx_hash"`
	NextTxHash string `gorm:"type:varchar(64);unique_index:idx_name_next_tx_hash"`
	CreatedAt  int64
}
