package scheduler

// Service defines a service
type Service struct {
	ID   string
	IP   string
	Port string
}

// ServiceDiscovery defines an interface for
// finding services within a cluster
type ServiceDiscovery interface {
	GetByID(id string) []Service
}
