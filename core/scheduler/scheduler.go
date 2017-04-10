package scheduler

// DeploymentInfo represents a successful deployment
type DeploymentInfo struct {
	ID     string
	EvalID string
}

// Scheduler defines an interface for cluster interactions
type Scheduler interface {
	GetName() string
	Deploy(jobID, lang, url, tag, buildParams, link, memory, cpuShare string) (*DeploymentInfo, error)
	SetAddr(addr string, https bool)
	GetServiceDiscoverer() ServiceDiscovery
	GetDeploymentStatus(jobID string) (string, error)
	Stop(jobID string) error
}
