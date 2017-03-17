package types

import "github.com/ellcrys/util"

// GlobalLedgerName represents the name of the global ledger
var globalLedgerName = util.Sha256("global")

// GetGlobalLedgerName returns the global ledger name
func GetGlobalLedgerName() string {
	return globalLedgerName
}

// Ledger represents a group of transactions
type Ledger struct {
	Number    uint   `gorm:"primary_key"`
	Hash      string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	Name      string `json:"name" gorm:"type:varchar(128);unique_index:idx_name_name"`
	Public    bool   `json:"public"`
	Chained   bool   `json:"chained"`
	CreatedAt int64  `json:"createdAt" gorm:"index:idx_name_created_at"`
}
