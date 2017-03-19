package types

import "github.com/ellcrys/util"

// globalChainName represents the name of the global ledger
var globalChainName = util.Sha256("global")

// GetGlobalChainName returns the global chain name
func GetGlobalChainName() string {
	return globalChainName
}

// Chain represents a table of chains
type Chain struct {
	Number    uint   `gorm:"primary_key"`
	Name      string `json:"name" gorm:"type:varchar(128);unique_index:idx_name_chain_name"`
	Public    bool   `json:"public"`
	CreatedAt int64  `json:"createdAt"`
}

// ToJSON returns the json equivalent of this object
func (c *Chain) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}
