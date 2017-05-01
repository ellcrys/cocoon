package types

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/fatih/structs"
	"github.com/jinzhu/copier"
	"github.com/ncodes/cocoon/core/common/mapdiff"
)

// Cocoon represents a smart contract application
type Cocoon struct {
	IdentityID            string   `json:"identityId,omitempty" structs:"identityId,omitempty" mapstructure:"identityId"`
	ID                    string   `json:"id,omitempty" structs:"id,omitempty" mapstructure:"id"`
	Memory                int      `json:"memory,omitempty" structs:"memory,omitempty" mapstructure:"memory"`
	CPUShare              int      `json:"CPUShare,omitempty" structs:"CPUShare,omitempty" mapstructure:"CPUShare"`
	NumSignatories        int      `json:"numSignatories,omitempty" structs:"numSignatories,omitempty" mapstructure:"numSignatories"`
	SigThreshold          int      `json:"sigThreshold,omitempty" structs:"sigThreshold,omitempty" mapstructure:"sigThreshold"`
	Releases              []string `json:"release,omitempty" structs:"releases,omitempty" mapstructure:"releases"`
	Signatories           []string `json:"signatories,omitempty" structs:"signatories,omitempty" mapstructure:"signatories"`
	Status                string   `json:"status,omitempty" structs:"status,omitempty" mapstructure:"status"`
	LastDeployedReleaseID string   `json:"lastDeployedReleaseID,omitempty" structs:"lastDeployedReleaseID,omitempty" mapstructure:"lastDeployedReleaseID"`
	CreatedAt             string   `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt"`
}

// Difference returns the difference between the current cocoon and another cocoon
func (c *Cocoon) Difference(o Cocoon) [][]mapdiff.DiffValue {
	return mapdiff.NewMapDiff(mapdiff.Map{Value: c.ToMap(), Name: c.ID}, mapdiff.Map{Value: o.ToMap(), Name: o.ID}).Diff()
}

// Clone creates a clone of this object.
func (c *Cocoon) Clone() Cocoon {
	var clone Cocoon
	copier.Copy(&clone, c)
	return clone
}

// ToJSON returns the json equivalent of this object
func (c *Cocoon) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}

// ToMap returns the map equivalent of the object
func (c *Cocoon) ToMap() map[string]interface{} {
	return structs.New(c).Map()
}

// MakeCocoonKey constructs a cocoon key
func MakeCocoonKey(id string) string {
	return fmt.Sprintf("cocoon;%s", id)
}

// MakeCocoonEnvKey returns a key for storing a cocoon environment variable
func MakeCocoonEnvKey(cocoonID string) string {
	return fmt.Sprintf("cocoon_env;%s", cocoonID)
}
