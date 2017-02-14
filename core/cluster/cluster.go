package cluster

// Cluster defines an interface for cluster interactions
type Cluster interface {
	Deploy(lang, url, tag string) (string, error)
	SetAddr(addr string, https bool)
}
