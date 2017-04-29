package types

import (
	"fmt"

	"github.com/ellcrys/util"
)

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string   `structs:"ID" mapstructure:"ID"`
	CocoonID    string   `structs:"cocoonID" mapstructure:"cocoonID"`
	URL         string   `structs:"URL" mapstructure:"URL"`
	Version     string   `structs:"version" mapstructure:"version"`
	Language    string   `structs:"language" mapstructure:"language"`
	BuildParam  string   `structs:"buildParam" mapstructure:"buildParam"`
	Link        string   `structs:"link" mapstructure:"link"`
	SigApproved int      `structs:"sigApproved" mapstructure:"sigApproved"`
	SigDenied   int      `structs:"sigDenied" mapstructure:"sigDenied"`
	VotersID    []string `structs:"votersID" mapstructure:"votersID"`
	Firewall    Firewall `structs:"firewall" mapstructure:"firewall"`
	ACL         ACLMap   `structs:"acl" mapstructure:"acl"`
	Env         Env      `structs:"env" mapstructure:"env"`
	CreatedAt   string   `structs:"createdAt" mapstructure:"createdAt"`
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
