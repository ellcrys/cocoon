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
var lockedKeys = make(map[string]*LockValue)
var lockKeyMu = sync.Mutex{}

// LockValue represents a lock value
type LockValue struct {
	Session string
	Exp     time.Time
}

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

	if l.state["lock_session"].(string) == "" {
		l.state["lock_session"] = util.UUID4()
	}

	lockKey := l.state["key"].(string)

	lockKeyMu.Lock()
	defer lockKeyMu.Unlock()
	if lockValue, hasKey := lockedKeys[lockKey]; hasKey {
		if lockValue.Session != l.state["lock_session"].(string) {
			return types.ErrLockAlreadyAcquired
		}
	}

	lockedKeys[lockKey] = &LockValue{
		Session: l.state["lock_session"].(string),
		Exp:     time.Now().UTC().Add(LockTTL),
	}

	return nil
}

// Release release a lock on the key. Returns nil of lock
// on the current key does not exist
func (l *Lock) Release() error {
	lockKeyMu.Lock()
	defer lockKeyMu.Unlock()
	key := l.state["key"].(string)
	if lockVal, hasLock := lockedKeys[key]; hasLock {
		if lockVal.Session == l.state["lock_session"].(string) {
			delete(lockedKeys, key)
		}
	}
	return nil
}

// IsAcquirer checks whether this lock instance is the acquirer of the lock on a specific key
func (l *Lock) IsAcquirer() error {

	if len(l.state["key"].(string)) == 0 {
		return fmt.Errorf("key is not set")
	}

	key := l.state["key"].(string)
	if lockVal, hasLock := lockedKeys[key]; hasLock {
		if lockVal.Session == l.state["lock_session"].(string) {
			return nil
		}
	}

	return types.ErrLockNotAcquired
}

// GetState returns the lock state
func (l *Lock) GetState() map[string]interface{} {
	return l.state
}

// StartLockWatcher starts a go routine that checks the locked key list
// and removes expired locks. Returns a function to stop the lock watcher
func StartLockWatcher() func() {
	var stop = false
	go func() {
		for !stop {
			lockKeyMu.Lock()
			for k, lockVal := range lockedKeys {
				if time.Now().UTC().After(lockVal.Exp) {
					delete(lockedKeys, k)
				}
			}
			lockKeyMu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()
	return func() {
		stop = true
	}
}
