package connector

import "github.com/ncodes/cocoon/core/types"

// Spec represents all the required deployment information
// required by a connector to operate
type Spec struct {
	ID          string
	Cocoon      *types.Cocoon
	URL         string
	Version     string
	Lang        string
	DiskLimit   int64
	BuildParams string
	Link        string
	Memory      int64
	CPUShare    int64
	ReleaseID   string
	Release     *types.Release
}
