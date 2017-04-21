package types

// Lock defines an interface for implementing a lock mechanism.
type Lock interface {
	// Acquire acquires a lock on key
	Acquire(key string) error
	// Release release the lock held
	Release() error
	// IsAcquirer checks if the lock is still held
	IsAcquirer() bool
	// GetState returns the state associated with the lock instance
	GetState() map[string]interface{}
}
