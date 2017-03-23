package scheduler

// Service defines a service
type Service struct {
	Name string
	ID   string
	IP   string
	Port string
	Tags []string
}

// ServiceDiscovery defines an interface for
// finding services within a cluster
type ServiceDiscovery interface {
	GetByID(name string, query map[string]string) ([]*Service, error)
}
