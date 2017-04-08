package types

import "github.com/ellcrys/util"
import "strings"

// Identity defines a user
type Identity struct {
	Email          string
	Password       string
	Cocoons        []string
	ClientSessions []string
	CreatedAt      string
}

// NewIdentity creates a new Identity
func NewIdentity(email, password string) *Identity {
	return &Identity{
		Email:    email,
		Password: password,
	}
}

// GetID returns the hashed version of the email
func (i *Identity) GetID() string {
	return util.Sha256(strings.ToLower(strings.TrimSpace(i.Email)))
}

// ToJSON returns the json equivalent of this object
func (i *Identity) ToJSON() []byte {
	json, _ := util.ToJSON(i)
	return json
}
