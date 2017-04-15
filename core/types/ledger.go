package types

import (
	"fmt"

	"github.com/ellcrys/util"
)

var (
	// systemPublicLedgerName represents the name of the public system ledger ledger
	systemPublicLedgerName = "public"

	// systemPrivateLedgerName represents the name of the private system ledger
	systemPrivateLedgerName = "private"
)

// GetSystemPublicLedgerName returns the systems public ledger name
func GetSystemPublicLedgerName() string {
	return systemPublicLedgerName
}

// GetSystemPrivateLedgerName returns the system's private ledger name
func GetSystemPrivateLedgerName() string {
	return systemPrivateLedgerName
}

// Ledger represents a group of transactions
type Ledger struct {
	Number       uint   `gorm:"primary_key"`
	Hash         string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	Name         string `json:"name" gorm:"type:varchar(128);unique_index:idx_name_name"`
	NameInternal string `json:"nameInternal" gorm:"-" sql:"-"`
	Public       bool   `json:"public"`
	Chained      bool   `json:"chained"`
	CreatedAt    int64  `json:"createdAt" gorm:"index:idx_name_created_at"`
}

// ToJSON returns the json equivalent of this object
func (l *Ledger) ToJSON() []byte {
	json, _ := util.ToJSON(l)
	return json
}

// MakeLedgerName creates a ledger name for use for creating or querying a ledger.
// Accepts a namespace value and the ledger name.
func MakeLedgerName(namespace, name string) string {
	return fmt.Sprintf("%s;%s", namespace, name)
}
