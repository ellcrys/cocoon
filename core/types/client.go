package types

import "github.com/ellcrys/util"

// Cocoon represents a smart contract application
type Cocoon struct {
	Identity       string
	ID             string
	URL            string
	ReleaseTag     string
	Language       string
	BuildParam     string
	Memory         string
	CPUShare       string
	Instances      int32
	NumSignatories int32
	SigThreshold   int32
	Releases       []string
	Signatories    []string
}

// ToJSON returns the json equivalent of this object
func (c *Cocoon) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string
	CocoonID    string
	URL         string
	ReleaseTag  string
	Language    string
	BuildParam  string
	SigApproved int32
	SigDenied   int32
}

// ToJSON returns the json equivalent of this object
func (r *Release) ToJSON() []byte {
	json, _ := util.ToJSON(r)
	return json
}
