package memory

import (
	"sync"
	"time"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/types"
)

// LockTTL defines max time to live of a lock.
var LockTTL = time.Duration(10 * time.Second)

// Global store for locked keys
var lockedKeys map[string]int64
var lockKeyMu = sync.Mutex{}

// Lock provides lock functionalities backed by memory.
type Lock struct {
	state map[string]interface{}
}

// NewLock creates a consul lock instance
func NewLock(key string) *Lock {
	return &Lock{
		state: map[string]interface{}{
			"lock_key_prefix": "platform/lock",
			"key":             key,
			"lock_session":    "",
		},
	}
}

// Acquire acquires a lock. A time-to-live time is set
// on the lock to ensure the lock is invalidated after the time is passed.
func (l *Lock) Acquire() error {
	if _, ok := lockedKeys[l.state["key"].(string)]; !ok {
		return types.ErrLockAlreadyAcquired
	}
	l.state["lock_session"] = util.UUID4()
	lockKey := fmt.Sprintf("%s.%s", l.state["lock_session"].(string), l.state["key"].(string))
	if _, ok := l.state[lockKey]; ok {
		return types.ErrLockAlreadyAcquired
	}
	lockedKeys[lockKey] = time.Now().UTC().Add(LockTTL).Unix()
	return nil
}

// Release release a lock on the key
func (l *Lock) Release() error {
	lockKeyMu.Lock()
	delete(lockedKeys, fmt.Sprintf("%s.%s", l.state["lock_session"].(string), l.state["key"].(string)))
	lockKeyMu.Unlock()
	return nil
}

// IsAcquirer checks whether this lock instance is the acquirer of the lock on a specific key
func (l *Lock) IsAcquirer() error {
	if len(l.state["key"].(string)) == 0 {
		return fmt.Errorf("key is not set")
	}

	lockKey := fmt.Sprintf("%s.%s", l.state["lock_session"].(string), l.state["key"].(string))
	if _, ok := l.state[lockKey]; !ok {
		return types.ErrLockNotAcquired
	}

	return nil
}

// GetState returns the lock state
func (l *Lock) GetState() map[string]interface{} {
	return l.state
}
