package types

// Ledger represents a group of linked transactions
type Ledger struct {
	Number         uint   `gorm:"primary_key"`
	Hash           string `gorm:"type:varchar(64);unique_index:idx_name_hash"`
	PrevLedgerHash string `gorm:"type:varchar(64);unique_index:idx_name_prev_ledger_hash"`
	NextLedgerHash string `gorm:"type:varchar(64);unique_index:idx_name_next_ledger_hash"`
	Name           string `gorm:"type:varchar(64);unique_index:idx_name_name"`
	Public         bool
	CreatedAt      int64
}
