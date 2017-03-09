package client

// Cocoon represents a cocoon
type Cocoon struct {
	URL        string
	ReleaseTag string
	Lang       string
	BuildParam string
	Memory     string
	CPUShare   string
	Instances  int32
}
