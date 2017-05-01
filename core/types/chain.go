package types

import "github.com/ellcrys/util"

// Chain represents a table of chains
type Chain struct {
	Number    uint   `json:"number,omitempty" structs:"number,omitempty" mapstructure:"number" gorm:"primary_key,omitempty"`
	Name      string `json:"name,omitempty" structs:"name,omitempty" mapstructure:"name" gorm:"type:varchar(128);unique_index:idx_name_chain_name,omitempty"`
	Public    bool   `json:"public,omitempty" structs:"public,omitempty" mapstructure:"public,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
}

// ToJSON returns the json equivalent of this object
func (c *Chain) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}
