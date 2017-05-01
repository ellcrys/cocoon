package types

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/fatih/structs"
	"github.com/imdario/mergo"
	"github.com/ncodes/cocoon/core/common/mapdiff"
	"github.com/ncodes/mapstructure"
)

// Cocoon represents a smart contract application
type Cocoon struct {
	IdentityID            string   `json:"identityId,omitempty" structs:"identityId,omitempty" mapstructure:"identityId,omitempty"`
	ID                    string   `json:"id,omitempty" structs:"id,omitempty" mapstructure:"id,omitempty"`
	Memory                int      `json:"memory,omitempty" structs:"memory,omitempty" mapstructure:"memory,omitempty"`
	CPUShare              int      `json:"CPUShare,omitempty" structs:"CPUShare,omitempty" mapstructure:"CPUShare,omitempty"`
	NumSignatories        int      `json:"numSignatories,omitempty" structs:"numSignatories,omitempty" mapstructure:"numSignatories,omitempty"`
	SigThreshold          int      `json:"sigThreshold,omitempty" structs:"sigThreshold,omitempty" mapstructure:"sigThreshold,omitempty"`
	Releases              []string `json:"release,omitempty" structs:"releases,omitempty" mapstructure:"releases,omitempty"`
	Signatories           []string `json:"signatories,omitempty" structs:"signatories,omitempty" mapstructure:"signatories,omitempty"`
	Status                string   `json:"status,omitempty" structs:"status,omitempty" mapstructure:"status,omitempty"`
	LastDeployedReleaseID string   `json:"lastDeployedReleaseID,omitempty" structs:"lastDeployedReleaseID,omitempty" mapstructure:"lastDeployedReleaseID,omitempty"`
	CreatedAt             string   `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
}

// Difference returns the difference between the current cocoon and another cocoon
func (c *Cocoon) Difference(o Cocoon) [][]mapdiff.DiffValue {
	return mapdiff.NewMapDiff(mapdiff.Map{Value: c.ToMap(), Name: c.ID}, mapdiff.Map{Value: o.ToMap(), Name: o.ID}).Diff()
}

// Clone creates a clone of this object.
func (c *Cocoon) Clone() Cocoon {
	var clone Cocoon
	m := c.ToMap()
	mapstructure.Decode(m, &clone)
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

// ToMapPtr same as ToMap but returns a pointer
func (c *Cocoon) ToMapPtr() *map[string]interface{} {
	ptr := structs.New(c).Map()
	return &ptr
}

// Merge merges another cocoon or map with current object replacing every
// non-empty value with non-empty values of the passed object.
func (c *Cocoon) Merge(o interface{}) error {
	switch obj := o.(type) {
	case Cocoon:
		m := c.ToMapPtr()
		mergo.MergeWithOverwrite(m, obj.ToMap())
		mapstructure.Decode(m, c)
	case map[string]interface{}:
		m := c.ToMapPtr()
		mergo.MergeWithOverwrite(m, obj)
		mapstructure.Decode(m, c)
	default:
		return fmt.Errorf("unsupported type")
	}
	return nil
}

// MakeCocoonKey constructs a cocoon key
func MakeCocoonKey(id string) string {
	return fmt.Sprintf("cocoon;%s", id)
}

// MakeCocoonEnvKey returns a key for storing a cocoon environment variable
func MakeCocoonEnvKey(cocoonID string) string {
	return fmt.Sprintf("cocoon_env;%s", cocoonID)
}
