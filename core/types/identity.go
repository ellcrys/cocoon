package types

import (
	"fmt"
	"strings"

	"github.com/ellcrys/util"
)

// Identity defines a user
type Identity struct {
	Email          string   `json:"email,omitempty" structs:"email,omitempty" mapstructure:"email,omitempty"`
	Password       string   `json:"password,omitempty" structs:"password,omitempty" mapstructure:"password,omitempty"`
	Cocoons        []string `json:"cocoons,omitempty" structs:"cocoons,omitempty" mapstructure:"cocoons,omitempty"`
	ClientSessions []string `json:"clientSessions,omitempty" structs:"clientSessions,omitempty" mapstructure:"clientSessions,omitempty"`
	CreatedAt      string   `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
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

// MakeIdentityKey constructs an identity key
func MakeIdentityKey(id string) string {
	return fmt.Sprintf("identity;%s", id)
}

// MakeIdentityPasswordKey returns a key for storing an identities password
func MakeIdentityPasswordKey(identityID string) string {
	return fmt.Sprintf("identity_password;%s", identityID)
}
