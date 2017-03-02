package types

// Ledger represents a group of linked transactions
type Ledger struct {
	Number         uint   `gorm:"primary_key"`
	Hash           string `json:"hash" gorm:"type:varchar(64);unique_index"`
	PrevLedgerHash string `json:"prev_ledger_hash" gorm:"type:varchar(64);unique_index"`
	NextLedgerHash string `json:"next_ledger_hash" gorm:"type:varchar(64);unique_index"`
	Name           string `json:"name" gorm:"type:varchar(64);unique_index"`
	Public         bool   `json:"public"`
	CreatedAt      int64  `json:"created_at"`
}
