package types

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/fatih/structs"
	"github.com/imdario/mergo"
	"github.com/ncodes/cocoon/core/common/mapdiff"
	"github.com/ncodes/mapstructure"
)

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string   `json:"id,omitempty" structs:"id,omitempty" mapstructure:"id,omitempty"`
	CocoonID    string   `json:"cocoonID,omitempty" structs:"cocoonID,omitempty" mapstructure:"cocoonID,omitempty"`
	URL         string   `json:"URL,omitempty" structs:"URL,omitempty" mapstructure:"URL,omitempty"`
	Version     string   `json:"version,omitempty" structs:"version,omitempty" mapstructure:"version,omitempty"`
	Language    string   `json:"language,omitempty" structs:"language,omitempty" mapstructure:"language,omitempty"`
	BuildParam  string   `json:"buildParam,omitempty" structs:"buildParam,omitempty" mapstructure:"buildParam,omitempty"`
	Link        string   `json:"link,omitempty" structs:"link,omitempty" mapstructure:"link,omitempty"`
	SigApproved int      `json:"sigApproved,omitempty" structs:"sigApproved,omitempty" mapstructure:"sigApproved,omitempty"`
	SigDenied   int      `json:"sigDenied,omitempty" structs:"sigDenied,omitempty" mapstructure:"sigDenied,omitempty"`
	VotersID    []string `json:"votersID,omitempty" structs:"votersID,omitempty" mapstructure:"votersID,omitempty"`
	Firewall    Firewall `json:"firewall,omitempty" structs:"firewall,omitempty" mapstructure:"firewall,omitempty"`
	ACL         ACLMap   `json:"acl,omitempty" structs:"acl,omitempty" mapstructure:"acl,omitempty"`
	Env         Env      `json:"env,omitempty" structs:"env,omitempty" mapstructure:"env,omitempty"`
	CreatedAt   string   `json:"createdAt,omitempty" structs:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
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

// ToMapPtr same as ToMap but returns a pointer
func (r *Release) ToMapPtr() *map[string]interface{} {
	ptr := structs.New(r).Map()
	return &ptr
}

// Merge merges another release or map with current object replacing every
// non-empty value with non-empty values of the passed object.
func (r *Release) Merge(o interface{}) error {
	switch obj := o.(type) {
	case Release:
		m := r.ToMapPtr()
		mergo.MergeWithOverwrite(m, obj.ToMap())
		mapstructure.Decode(m, r)
	case map[string]interface{}:
		m := r.ToMapPtr()
		mergo.MergeWithOverwrite(m, obj)
		mapstructure.Decode(m, r)
	default:
		return fmt.Errorf("unsupported type")
	}
	return nil
}

// Clone creates a clone of this object
func (r *Release) Clone() Release {
	var clone Release
	m := r.ToMap()
	mapstructure.Decode(m, &clone)
	return clone
}

// MakeReleaseKey constructs an release key
func MakeReleaseKey(id string) string {
	return fmt.Sprintf("release;%s", id)
}

// MakePrivateReleaseKey returns a key for storing private release fields
func MakePrivateReleaseKey(releaseID string) string {
	return fmt.Sprintf("release_private;%s", releaseID)
}
