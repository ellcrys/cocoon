package connector

// Request defines a launch request
type Request struct {
	ID          string
	URL         string
	Version         string
	Lang        string
	DiskLimit   int64
	BuildParams string
	Link        string
	Memory      int64
	CPUShare   int64
}
