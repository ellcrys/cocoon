package consul

import (
	"fmt"
	"time"

	"strings"

	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/types"
)

// LockTTL defines max time to live of a lock.
var LockTTL = time.Duration(10 * time.Second)

func init() {
	goreq.SetConnectTimeout(time.Second * 10)
}

// Lock provides lock functionalities based on consul sessions.It implements
// The Lock interface.
type Lock struct {
	lockTTL time.Duration
	state   map[string]interface{}
}

// NewLock creates a consul lock instance
func NewLock(key string) *Lock {
	return &Lock{
		lockTTL: LockTTL,
		state: map[string]interface{}{
			"consul_addr":     "http://localhost:8500",
			"lock_key_prefix": "platform/lock",
			"key":             key,
			"lock_session":    "",
		},
	}
}

// NewLockWithTTL creates a consul lock instance
func NewLockWithTTL(key string, ttl time.Duration) *Lock {
	return &Lock{
		lockTTL: ttl,
		state: map[string]interface{}{
			"consul_addr":     "http://localhost:8500",
			"lock_key_prefix": "platform/lock",
			"key":             key,
			"lock_session":    "",
		},
	}
}

// createSession creates a consul session
func (l *Lock) createSession(ttl int) (string, error) {
	var ttlStr string
	if ttl > 0 {
		ttlStr = fmt.Sprintf("%ds", ttl)
	}
	resp, err := goreq.Request{
		Method: "PUT",
		Uri:    l.state["consul_addr"].(string) + "/v1/session/create",
		Body: map[string]string{
			"TTL":       ttlStr,
			"Behaviour": "delete",
			"LockDelay": "5s",
		},
	}.Do()
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := resp.Body.ToString()
		return "", fmt.Errorf(b)
	}

	var result map[string]string
	if err = resp.Body.FromJsonTo(&result); err != nil {
		return "", err
	}

	return result["ID"], nil
}

func (l *Lock) acquire() error {
	resp, err := goreq.Request{
		Method: "PUT",
		Uri:    l.state["consul_addr"].(string) + "/v1/kv/" + fmt.Sprintf("%s.%s?acquire=%s", l.state["lock_key_prefix"].(string), l.state["key"].(string), l.state["lock_session"].(string)),
	}.Do()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := resp.Body.ToString()
		return fmt.Errorf(b)
	}

	status, _ := resp.Body.ToString()
	if strings.TrimSpace(status) == "false" {
		return types.ErrLockAlreadyAcquired
	}

	return nil
}

// Acquire acquires a lock. A time-to-live time is set
// on the lock to ensure the lock is invalidated after the time is passed.
func (l *Lock) Acquire() error {
	var err error

	// If lock object has got a session, get one.
	if l.state["lock_session"].(string) == "" {
		l.state["lock_session"], err = l.createSession(int(l.lockTTL.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to get lock: %s", err)
		}
	}
	return l.acquire()
}

// Release invalidates the lock previously acquired
func (l *Lock) Release() error {
	resp, err := goreq.Request{
		Method: "PUT",
		Uri:    l.state["consul_addr"].(string) + "/v1/kv/" + fmt.Sprintf("%s.%s?release=%s", l.state["lock_key_prefix"].(string), l.state["key"].(string), l.state["lock_session"].(string)),
	}.Do()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := resp.Body.ToString()
		return fmt.Errorf(b)
	}

	return nil
}

// IsAcquirer checks whether this lock instance is the acquirer of the lock on a specific key
func (l *Lock) IsAcquirer() error {
	if len(l.state["key"].(string)) == 0 {
		return fmt.Errorf("key is not set")
	}

	resp, err := goreq.Request{
		Uri: l.state["consul_addr"].(string) + "/v1/kv/" + fmt.Sprintf("%s.%s", l.state["lock_key_prefix"].(string), l.state["key"].(string)),
	}.Do()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return types.ErrLockNotAcquired
	}

	var sessions []map[string]interface{}
	err = resp.Body.FromJsonTo(&sessions)
	if err != nil {
		return err
	}

	if len(sessions) > 0 {
		session := sessions[0]["Session"]
		if session == nil {
			return types.ErrLockNotAcquired
		}
		if session.(string) != l.state["lock_session"].(string) {
			return types.ErrLockNotAcquired
		}
	} else {
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
