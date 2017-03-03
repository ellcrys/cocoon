package types

// Ledger represents a group of linked transactions
type Ledger struct {
	Number         uint   `gorm:"primary_key"`
	Hash           string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	PrevLedgerHash string `json:"prevLedgerHash" gorm:"type:varchar(64);unique_index:idx_name_prev_ledger_hash"`
	NextLedgerHash string `json:"nextLedgerHash" gorm:"type:varchar(64);unique_index:idx_name_next_ledger_hash"`
	Name           string `json:"name" gorm:"type:varchar(64);unique_index:idx_name_name"`
	Public         bool   `json:"public"`
	CreatedAt      int64  `json:"createdAt"`
}
