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
	Number       uint   `json:"number,omitempty" structs:"number,omitempty" mapstructure:"number" gorm:"primary_key,omitempty"`
	Name         string `json:"name,omitempty" structs:"name,omitempty" mapstructure:"name" gorm:"type:varchar(128);unique_index:idx_name_name,omitempty"`
	NameInternal string `json:"-" structs:"nameInternal,omitempty" mapstructure:"nameInternal" gorm:"-" sql:"-,omitempty"`
	CocoonID     string `json:"cocoonId,omitempty" structs:"cocoonId,omitempty" mapstructure:"cocoonId" gorm:"index:idx_name_cocoon_id,omitempty"`
	Public       bool   `json:"public,omitempty" structs:"public,omitempty" mapstructure:"public" json:"public,omitempty"`
	Chained      bool   `json:"chained,omitempty" structs:"chained,omitempty" mapstructure:"chained,omitempty"`
	CreatedAt    int64  `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt" gorm:"index:idx_name_created_at,omitempty"`
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
