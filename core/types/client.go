package types

import "github.com/ellcrys/util"

// Cocoon represents a smart contract application
type Cocoon struct {
	IdentityID     string   `structs:"identity_id" mapstructure:"identity_id"`
	ID             string   `structs:"ID" mapstructure:"ID"`
	URL            string   `structs:"URL" mapstructure:"URL"`
	ReleaseTag     string   `structs:"releaseTag" mapstructure:"releaseTag"`
	Language       string   `structs:"language" mapstructure:"language"`
	BuildParam     string   `structs:"buildParam" mapstructure:"buildParam"`
	Memory         string   `structs:"memory" mapstructure:"memory"`
	CPUShares      string   `structs:"CPUShares" mapstructure:"CPUShares"`
	NumSignatories int32    `structs:"numSignatories" mapstructure:"numSignatories"`
	SigThreshold   int32    `structs:"sigThreshold" mapstructure:"sigThreshold"`
	Link           string   `structs:"link" mapstructure:"link"`
	Releases       []string `structs:"releases" mapstructure:"releases"`
	Signatories    []string `structs:"signatories" mapstructure:"signatories"`
	CreatedAt      string   `structs:"createdAt" mapstructure:"createdAt"`
}

// ToJSON returns the json equivalent of this object
func (c *Cocoon) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string   `structs:"ID" mapstructure:"ID"`
	CocoonID    string   `structs:"cocoonID" mapstructure:"cocoonID"`
	URL         string   `structs:"URL" mapstructure:"URL"`
	ReleaseTag  string   `structs:"releaseTag" mapstructure:"releaseTag"`
	Language    string   `structs:"language" mapstructure:"language"`
	BuildParam  string   `structs:"buildParam" mapstructure:"buildParam"`
	Link        string   `structs:"link" mapstructure:"link"`
	SigApproved int32    `structs:"sigApproved" mapstructure:"sigApproved"`
	SigDenied   int32    `structs:"sigDenied" mapstructure:"sigDenied"`
	VotersID    []string `structs:"votersID" mapstructure:"votersID"`
	CreatedAt   string   `structs:"createdAt" mapstructure:"createdAt"`
}

// ToJSON returns the json equivalent of this object
func (r *Release) ToJSON() []byte {
	json, _ := util.ToJSON(r)
	return json
}
