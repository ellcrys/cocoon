package types

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/fatih/structs"
	"github.com/jinzhu/copier"
	"github.com/ncodes/cocoon/core/common/mapdiff"
)

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string   `json:"id,omitempty" structs:"id,omitempty" mapstructure:"id"`
	CocoonID    string   `json:"cocoonID,omitempty" structs:"cocoonID,omitempty" mapstructure:"cocoonID"`
	URL         string   `json:"URL,omitempty" structs:"URL,omitempty" mapstructure:"URL"`
	Version     string   `json:"version,omitempty" structs:"version,omitempty" mapstructure:"version"`
	Language    string   `json:"language,omitempty" structs:"language,omitempty" mapstructure:"language"`
	BuildParam  string   `json:"buildParam,omitempty" structs:"buildParam,omitempty" mapstructure:"buildParam"`
	Link        string   `json:"link,omitempty" structs:"link,omitempty" mapstructure:"link"`
	SigApproved int      `json:"sigApproved,omitempty" structs:"sigApproved,omitempty" mapstructure:"sigApproved"`
	SigDenied   int      `json:"sigDenied,omitempty" structs:"sigDenied,omitempty" mapstructure:"sigDenied"`
	VotersID    []string `json:"votersID,omitempty" structs:"votersID,omitempty" mapstructure:"votersID"`
	Firewall    Firewall `json:"firewall,omitempty" structs:"firewall,omitempty" mapstructure:"firewall"`
	ACL         ACLMap   `json:"acl,omitempty" structs:"acl,omitempty" mapstructure:"acl"`
	Env         Env      `json:"env,omitempty" structs:"env,omitempty" mapstructure:"env"`
	CreatedAt   string   `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt"`
}

// Difference returns the difference between the current release and another release
func (r *Release) Difference(o Release) [][]mapdiff.DiffValue {
	return mapdiff.NewMapDiff(mapdiff.Map{Value: r.ToMap(), Name: r.ID}, mapdiff.Map{Value: o.ToMap(), Name: o.ID}).Diff()
}

// ToJSON returns the json equivalent of this object
func (r *Release) ToJSON() []byte {
	json, _ := util.ToJSON(r)
	return json
}

// ToMap returns the map equivalent of the object
func (r *Release) ToMap() map[string]interface{} {
	return structs.New(r).Map()
}

// Clone creates a clone of this object
func (r *Release) Clone() Release {
	var clone Release
	copier.Copy(&clone, r)
	return clone
}

// MakeReleaseKey constructs an release key
func MakeReleaseKey(id string) string {
	return fmt.Sprintf("release;%s", id)
}

// MakeReleaseEnvKey returns a key for storing a release environment variable
func MakeReleaseEnvKey(releaseID string) string {
	return fmt.Sprintf("release_env;%s", releaseID)
}
