package types

import (
	"fmt"

	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
)

// Transaction reprents a group of transactions belonging to a ledger.
// All transaction entries must reference the hash of the immediate transaction
// sharing the same ledger name.
type Transaction struct {
	Number    uint   `json:"number,omitempty" gorm:"primary_key"`
	Ledger    string `json:"ledger" gorm:"type:varchar(128);index:idx_name_ledger_name"`
	ID        string `json:"id" gorm:"type:varchar(64);unique_index:idx_name_id"`
	Key       string `json:"key" gorm:"type:varchar(128);index:idx_name_key"`
	Value     string `json:"value" gorm:"type:text"`
	Hash      string `json:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	BlockID   string `json:"blockId,omitempty"`
	Block     *Block `json:"block,omitempty" gorm:"-" sql:"-"`
	CreatedAt int64  `json:"createdAt" gorm:"index:idx_name_created_at"`
}

// MakeHash creates a hash of a transaction
func (t *Transaction) MakeHash() string {
	return util.Sha256(fmt.Sprintf(
		"%s %s %s %d",
		t.ID,
		crypto.ToBase64([]byte(t.Key)),
		crypto.ToBase64([]byte(t.Value)),
		t.CreatedAt))
}

// ToJSON returns the json equivalent of this object
func (t *Transaction) ToJSON() []byte {
	json, _ := util.ToJSON(t)
	return json
}
