package scheduler

// DeploymentInfo represents a successful deployment
type DeploymentInfo struct {
	ID     string
	EvalID string
}

// Service defines a service
type Service struct {
	ID   string
	IP   string
	Port string
}

// Scheduler defines an interface for cluster interactions
type Scheduler interface {
	Deploy(jobID, lang, url, tag, buildParams, link, memory, cpuShare string) (*DeploymentInfo, error)
	SetAddr(addr string, https bool)
	GetServices(serviceID string) []Service
}
