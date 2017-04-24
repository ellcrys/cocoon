package types

// Lock defines an interface for implementing a lock mechanism.
type Lock interface {
	// Acquire acquires a lock on key
	Acquire() error
	// Release release the lock held
	Release() error
	// IsAcquirer checks if the lock is still held
	IsAcquirer() error
	// GetState returns the state associated with the lock instance
	GetState() map[string]interface{}
	// SetState loads a lock state into the lock object
	SetState(map[string]interface{})
}
