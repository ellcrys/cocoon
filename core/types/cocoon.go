package types

import (
	"fmt"
	"reflect"

	"github.com/ellcrys/util"
	"github.com/fatih/structs"
)

// Cocoon represents a smart contract application
type Cocoon struct {
	IdentityID            string   `structs:"identity_id,omitempty" mapstructure:"identity_id"`
	ID                    string   `structs:"ID,omitempty" mapstructure:"ID"`
	Memory                int      `structs:"memory,omitempty" mapstructure:"memory"`
	CPUShare              int      `structs:"CPUShare,omitempty" mapstructure:"CPUShare"`
	NumSignatories        int      `structs:"numSignatories,omitempty" mapstructure:"numSignatories"`
	SigThreshold          int      `structs:"sigThreshold,omitempty" mapstructure:"sigThreshold"`
	Releases              []string `structs:"releases,omitempty" mapstructure:"releases"`
	Signatories           []string `structs:"signatories,omitempty" mapstructure:"signatories"`
	Status                string   `structs:"status,omitempty" mapstructure:"status"`
	LastDeployedReleaseID string   `structs:"lastDeployedReleaseID,omitempty" mapstructure:"lastDeployedReleaseID"`
	LastDeployedRelease   *Release `structs:"lastDeployedRelease,omitempty" mapstructure:"lastDeployedRelease" json:"-"`
	CreatedAt             string   `structs:"createdAt,omitempty" mapstructure:"createdAt"`
}

// HasFieldsChanged takes a set of fields and checks whether they have the same
// values as those of same fields on the object. It returns false as soon as it finds
// the first element that exists in the object and has a different value.
func (c *Cocoon) HasFieldsChanged(fields map[string]interface{}) bool {
	changed := false
	cMap := structs.New(c).Map()
	for oKey, oVal := range fields {
		if cVal, ok := cMap[oKey]; ok {
			if !reflect.DeepEqual(cVal, oVal) {
				changed = true
				break
			}
		}
	}
	return changed
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
