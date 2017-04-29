package types

import (
	"fmt"

	"github.com/ellcrys/util"
)

// Cocoon represents a smart contract application
type Cocoon struct {
	IdentityID          string   `structs:"identity_id" mapstructure:"identity_id"`
	ID                  string   `structs:"ID" mapstructure:"ID"`
	URL                 string   `structs:"URL" mapstructure:"URL"`
	Version             string   `structs:"version" mapstructure:"version"`
	Language            string   `structs:"language" mapstructure:"language"`
	BuildParam          string   `structs:"buildParam" mapstructure:"buildParam"`
	Memory              int      `structs:"memory" mapstructure:"memory"`
	CPUShare            int      `structs:"CPUShare" mapstructure:"CPUShare"`
	NumSignatories      int      `structs:"numSignatories" mapstructure:"numSignatories"`
	SigThreshold        int      `structs:"sigThreshold" mapstructure:"sigThreshold"`
	Link                string   `structs:"link" mapstructure:"link"`
	Releases            []string `structs:"releases" mapstructure:"releases"`
	Signatories         []string `structs:"signatories" mapstructure:"signatories"`
	Status              string   `structs:"status" mapstructure:"status"`
	LastDeployedRelease string   `structs:"lastDeployedRelease" mapstructure:"lastDeployedRelease"`
	ACL                 ACLMap   `structs:"ACL" mapstructure:"ACL"`
	Firewall            Firewall `structs:"firewall" mapstructure:"firewall"`
	Env                 Env      `structs:"env" mapstructure:"env"`
	CreatedAt           string   `structs:"createdAt" mapstructure:"createdAt"`
}

// ToJSON returns the json equivalent of this object
func (c *Cocoon) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}

// MakeCocoonKey constructs a cocoon key
func MakeCocoonKey(id string) string {
	return fmt.Sprintf("cocoon;%s", id)
}

// MakeCocoonEnvKey returns a key for storing a cocoon environment variable
func MakeCocoonEnvKey(cocoonID string) string {
	return fmt.Sprintf("cocoon_env;%s", cocoonID)
}
