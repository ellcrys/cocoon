package types

import (
	"fmt"
	"strings"

	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
)

// Transaction reprents a group of transactions belonging to a ledger.
// All transaction entries must reference the hash of the immediate transaction
// sharing the same ledger name.
type Transaction struct {
	Number         uint   `json:"number,omitempty" structs:"number,omitempty" mapstructure:"number" gorm:"primary_key"`
	Ledger         string `json:"ledger,omitempty" structs:"ledger,omitempty" mapstructure:"ledger" gorm:"type:varchar(128);index:idx_name_ledger_name"`
	ID             string `json:"id,omitempty" structs:"id,omitempty" mapstructure:"id" gorm:"type:varchar(64);unique_index:idx_name_id"`
	Key            string `json:"key,omitempty" structs:"key,omitempty" mapstructure:"key" gorm:"type:varchar(128);index:idx_name_key"`
	Value          string `json:"value,omitempty" structs:"value,omitempty" mapstructure:"value" gorm:"type:text"`
	Hash           string `json:"hash,omitempty" structs:"hash,omitempty" mapstructure:"hash" gorm:"type:varchar(64);unique_index:idx_name_hash"`
	BlockID        string `json:"blockId,omitempty" structs:"blockId,omitempty" mapstructure:"blockId"`
	RevisionTo     string `json:"revisionTo,omitempty" structs:"revisionTo,omitempty" mapstructure:"revisionTo" gorm:"type:varchar(64);unique_index:idx_name_revision_to" sql:"DEFAULT:NULL"`
	CreatedAt      int64  `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt" gorm:"index:idx_name_created_at"`
	LedgerInternal string `json:"-" structs:"-" mapstructure:"-" gorm:"-" sql:"-"`
	KeyInternal    string `json:"-" structs:"-" mapstructure:"-" gorm:"-" sql:"-"`
	Block          *Block `json:"block" structs:"block" mapstructure:"block" gorm:"-" sql:"-"`
}

// MakeHash creates a hash of a transaction
func (t *Transaction) MakeHash() string {
	return util.Sha256(fmt.Sprintf(
		"%s;%s;%s;%s;%s;%d",
		t.ID,
		crypto.ToBase64([]byte(t.LedgerInternal)),
		crypto.ToBase64([]byte(t.KeyInternal)),
		crypto.ToBase64([]byte(t.Value)),
		crypto.ToBase64([]byte(t.BlockID)),
		t.CreatedAt))
}

// ToJSON returns the json equivalent of this object
func (t *Transaction) ToJSON() []byte {
	json, _ := util.ToJSON(t)
	return json
}

// MakeTxKey creates a transaction key name for storing and querying a transaction.
// Accepts a namespace value and the key name.
func MakeTxKey(namespace, name string) string {
	return fmt.Sprintf("%s;%s", namespace, name)
}

// GetActualKeyFromTxKey returns the real key name from a transaction key
func GetActualKeyFromTxKey(key string) string {
	return strings.Split(key, ";")[1]
}

// TxReceipt defines a structure for representing transaction status
// from endpoint that manipulate transactions
type TxReceipt struct {
	ID  string `json:"id,omitempty"`
	Err string `json:"err,omitempty"`
}

// PutResult defines a structure for put operation ressult
type PutResult struct {
	TxReceipts []*TxReceipt `json:"txReceipts,omitempty"`
	Block      *Block       `json:"block,omitempty"`
}
