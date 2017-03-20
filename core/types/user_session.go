package types

import "github.com/ellcrys/util"

// UserSession represents a user session
type UserSession struct {
	Email string
	Token string
}

// ToJSON returns json encoding of instance
func (u *UserSession) ToJSON() []byte {
	json, _ := util.ToJSON(u)
	return json
}
