package scheduler

// DeploymentInfo represents a successful deployment
type DeploymentInfo struct {
	ID string
}

// Scheduler defines an interface for cluster interactions
type Scheduler interface {
	GetName() string
	Deploy(jobID, releaseID string, memory, cpuShare int) (*DeploymentInfo, error)
	SetAddr(addr string, https bool)
	GetServiceDiscoverer() (ServiceDiscovery, error)
	GetDeploymentStatus(jobID string) (string, error)
	Stop(jobID string) error
}
