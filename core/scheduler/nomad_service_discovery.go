package scheduler

// NomadServiceDiscovery provides service discovery to the nomad schedulerAddr
// by querying a consul server
type NomadServiceDiscovery struct {
	ConsulAddr string
	Protocol   string
}

// GetByID fetches a services instances by the service id.
func (nsd *NomadServiceDiscovery) GetByID(id string) []Service {
	log.Info("Hello")
	return nil
}
