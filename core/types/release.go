package types

import (
	"fmt"
	"reflect"

	"github.com/ellcrys/util"
	"github.com/fatih/structs"
)

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string   `structs:"ID,omitempty" mapstructure:"ID"`
	CocoonID    string   `structs:"cocoonID,omitempty" mapstructure:"cocoonID"`
	URL         string   `structs:"URL,omitempty" mapstructure:"URL"`
	Version     string   `structs:"version,omitempty" mapstructure:"version"`
	Language    string   `structs:"language,omitempty" mapstructure:"language"`
	BuildParam  string   `structs:"buildParam,omitempty" mapstructure:"buildParam"`
	Link        string   `structs:"link,omitempty" mapstructure:"link"`
	SigApproved int      `structs:"sigApproved,omitempty" mapstructure:"sigApproved"`
	SigDenied   int      `structs:"sigDenied,omitempty" mapstructure:"sigDenied"`
	VotersID    []string `structs:"votersID,omitempty" mapstructure:"votersID"`
	Firewall    Firewall `structs:"firewall,omitempty" mapstructure:"firewall"`
	ACL         ACLMap   `structs:"acl,omitempty" mapstructure:"acl"`
	Env         Env      `structs:"env,omitempty" mapstructure:"env"`
	CreatedAt   string   `structs:"createdAt,omitempty" mapstructure:"createdAt"`
}

// HasFieldsChanged takes a set of fields and checks whether they have the same
// values as those of same fields on the object. It returns false as soon as it finds
// the first element that exists in the object and has a different value.
func (r *Release) HasFieldsChanged(fields map[string]interface{}) bool {
	changed := false
	cMap := structs.New(r).Map()
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
func (r *Release) ToJSON() []byte {
	json, _ := util.ToJSON(r)
	return json
}

// MakeReleaseKey constructs an release key
func MakeReleaseKey(id string) string {
	return fmt.Sprintf("release;%s", id)
}

// MakeReleaseEnvKey returns a key for storing a release environment variable
func MakeReleaseEnvKey(releaseID string) string {
	return fmt.Sprintf("release_env;%s", releaseID)
}
