package connector

// Request defines a launch request
type Request struct {
	ID          string
	URL         string
	Tag         string
	Lang        string
	DiskLimit   int64
	BuildParams string
	Link        string
	CocoonAddr  string
}
