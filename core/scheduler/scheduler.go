package scheduler

// DeploymentInfo represents a successful deployment
type DeploymentInfo struct {
	ID     string
	EvalID string
}

// Scheduler defines an interface for cluster interactions
type Scheduler interface {
	Deploy(jobID, lang, url, tag, buildParams string) (*DeploymentInfo, error)
	SetAddr(addr string, https bool)
}
