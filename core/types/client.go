package types

import "github.com/ellcrys/util"

// Cocoon represents a smart contract application
type Cocoon struct {
	Identity       string   `structs:"identity" mapstructure:"identity"`
	ID             string   `structs:"ID" mapstructure:"ID"`
	URL            string   `structs:"URL" mapstructure:"URL"`
	ReleaseTag     string   `structs:"releaseTag "mapstructure:"releaseTag"`
	Language       string   `structs:"language" mapstructure:"language"`
	BuildParam     string   `structs:"buildParam" mapstructure:"buildParam"`
	Memory         string   `structs:"memory" mapstructure:"memory"`
	CPUShare       string   `structs:"CPUShare" mapstructure:"CPUShare"`
	Instances      int32    `structs:"instances" mapstructure:"instances"`
	NumSignatories int32    `structs:"numSignatories" mapstructure:"numSignatories"`
	SigThreshold   int32    `structs:"sigThreshold" mapstructure:"sigThreshold"`
	Releases       []string `structs:"releases" mapstructure:"releases"`
	Signatories    []string `structs:"signatories" mapstructure:"signatories"`
}

// ToJSON returns the json equivalent of this object
func (c *Cocoon) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string `structs:"ID" mapstructure:"ID"`
	CocoonID    string `structs:"cocoonID" mapstructure:"cocoonID"`
	URL         string `structs:"URL" mapstructure:"URL"`
	ReleaseTag  string `structs:"releaseTag" mapstructure:"releaseTag"`
	Language    string `structs:"language" mapstructure:"language"`
	BuildParam  string `structs:"buildParam" mapstructure:"buildParam"`
	SigApproved int32  `structs:"sigApproved" mapstructure:"sigApproved"`
	SigDenied   int32  `structs:"sigDenied" mapstructure:"sigDenied"`
}

// ToJSON returns the json equivalent of this object
func (r *Release) ToJSON() []byte {
	json, _ := util.ToJSON(r)
	return json
}
