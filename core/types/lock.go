package types

// Lock defines an interface for implementing a lock mechanism.
type Lock interface {
	Acquire(key string) error
	Release() error
	IsAcquirer(key string) bool
}
