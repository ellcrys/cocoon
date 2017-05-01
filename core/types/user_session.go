package types

import "github.com/ellcrys/util"

// UserSession represents a user session
type UserSession struct {
	Email string `json:"email,omitempty" structs:"email,omitempty" mapstructure:"email,omitempty"`
	Token string `json:"hash,omitempty" structs:"hash,omitempty" mapstructure:"hash,omitempty"`
}

// ToJSON returns json encoding of instance
func (u *UserSession) ToJSON() []byte {
	json, _ := util.ToJSON(u)
	return json
}
