package consul

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/hashicorp/consul/api"
	"github.com/ncodes/cocoon/core/types"
)

// LockTTL defines max time to live of a lock.
var LockTTL = 10 * time.Second

// WaitForLock defines the about of time to wait to acquire a lock
var WaitForLock = 5 * time.Second

func init() {
	goreq.SetConnectTimeout(time.Second * 10)
}

// Lock provides lock functionalities based on consul sessions.It implements
// The Lock interface.
type Lock struct {
	client  *api.Client
	lock    *api.Lock
	lockTTL time.Duration
	state   map[string]interface{}
}

// NewLock creates a consul lock instance
func NewLock(key string) (*Lock, error) {
	cfg := api.DefaultConfig()
	cfg.Address = util.Env("CONSUL_ADDR", cfg.Address)
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to reach consul server: %s", err)
	}
	return &Lock{
		client:  client,
		lockTTL: LockTTL,
		state: map[string]interface{}{
			"consul_addr":     "http://localhost:8500",
			"lock_key_prefix": "platform/lock",
			"key":             key,
			"lock_session":    "",
		},
	}, nil
}

// NewLockWithTTL creates a consul lock instance
func NewLockWithTTL(key string, ttl time.Duration) (*Lock, error) {
	cfg := api.DefaultConfig()
	cfg.Address = util.Env("CONSUL_ADDR", cfg.Address)
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to reach consul server: %s", err)
	}
	return &Lock{
		client:  client,
		lockTTL: ttl,
		state: map[string]interface{}{
			"consul_addr":     "http://localhost:8500",
			"lock_key_prefix": "platform/lock",
			"key":             key,
			"lock_session":    "",
		},
	}, nil
}

// createSession creates a consul session
func (l *Lock) createSession(ttl int) (string, error) {

	var ttlStr string
	if ttl > 0 {
		ttlStr = fmt.Sprintf("%ds", ttl)
	}

	session := l.client.Session()
	id, _, err := session.Create(&api.SessionEntry{
		TTL:       ttlStr,
		Behavior:  "delete",
		LockDelay: 5 * time.Second,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create error: %s", err)
	}

	return id, nil
}

// Acquire acquires a lock. Returns error if it failed to acquire the lock.
// If the lock object has a session that is still tired to the locked key,
// nil is returned.
func (l *Lock) Acquire() error {

	// If lock object has got a session, get one.
	if l.state["lock_session"].(string) == "" {
		sessionID, err := l.createSession(int(l.lockTTL.Seconds()))
		if err != nil {
			return err
		}
		l.state["lock_session"] = sessionID
	}

	lock, err := l.client.LockOpts(&api.LockOptions{
		Key:          l.makeLockKey(),
		Session:      l.state["lock_session"].(string),
		LockWaitTime: WaitForLock,
	})
	if err != nil {
		return err
	}

	// stop further lock retry attempt after WaitForLock seconds
	stopCh := make(chan struct{}, 1)
	time.AfterFunc(WaitForLock, func() {
		close(stopCh)
	})

	_, err = lock.Lock(stopCh)
	if err != nil {
		fmt.Println(">>>> ", err)
		return fmt.Errorf("failed to acquire lock: %s", err)
	}

	// check if lock is acquired
	if err = l.IsAcquirer(); err != nil {
		return err
	}

	l.lock = lock

	return nil
}

func (l *Lock) makeLockKey() string {
	lockKeyPrefix := l.state["lock_key_prefix"].(string)
	key := l.state["key"].(string)
	return fmt.Sprintf("%s/%s", lockKeyPrefix, key)
}

// Release invalidates the lock previously acquired.
// Returns nil if when lock is not held
func (l *Lock) Release() error {
	var err error

	if l.lock == nil {
		// If lock session is set, we need to recreate the lock
		// and attempt to acquire it. If we successfully acquired it, then we can proceed to releasing it.
		if len(l.state["lock_session"].(string)) > 0 {
			if err = l.Acquire(); err != nil {
				return nil
			}
		} else {
			return nil
		}
	}

	if err = l.lock.Unlock(); err != nil {
		if err.Error() == "Lock not held" {
			return nil
		}
		return fmt.Errorf("failed to release lock: %s", err)
	}

	return l.lock.Destroy()
}

// IsAcquirer checks whether this lock instance is the acquirer of the lock on a specific key
func (l *Lock) IsAcquirer() error {

	if len(l.state["key"].(string)) == 0 {
		return fmt.Errorf("key is not set")
	}

	key, _, err := l.client.KV().Get(l.makeLockKey(), nil)
	if err != nil {
		return err
	}

	if key == nil || key.Session != l.state["lock_session"].(string) {
		return types.ErrLockNotAcquired
	}

	return nil
}

// GetState returns the lock state
func (l *Lock) GetState() map[string]interface{} {
	return l.state
}

// SetState sets the lock state
func (l *Lock) SetState(state map[string]interface{}) {
	for k, s := range state {
		l.state[k] = s
	}
}
