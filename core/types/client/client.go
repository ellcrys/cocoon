package client

// Cocoon represents a smart contract application
type Cocoon struct {
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
