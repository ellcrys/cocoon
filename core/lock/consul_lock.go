package lock

import (
	"fmt"
	"time"

	"strings"

	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/types"
)

// LockTTL defines max time to live of a lock.
var LockTTL = time.Duration(10 * time.Second)

func init() {
	goreq.SetConnectTimeout(time.Second * 10)
}

// ConsulLock provides lock functionalities based on consul sessions.It implements
// The Lock interface.
type ConsulLock struct {
	consulAddr    string
	key           string
	lockKeyPrefix string
	lockSession   string
}

// NewConsulLock creates a consul lock instance
func NewConsulLock() *ConsulLock {
	return &ConsulLock{
		consulAddr:    "http://localhost:8500",
		lockKeyPrefix: "platform/lock",
	}
}

// createSession creates a consul session
func (l *ConsulLock) createSession(ttl int) (string, error) {
	var ttlStr string
	if ttl > 0 {
		ttlStr = fmt.Sprintf("%ds", ttl)
	}
	resp, err := goreq.Request{
		Method: "PUT",
		Uri:    l.consulAddr + "/v1/session/create",
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

func (l *ConsulLock) acquire(key string) error {
	resp, err := goreq.Request{
		Method: "PUT",
		Uri:    l.consulAddr + "/v1/kv/" + fmt.Sprintf("%s.%s?acquire=%s", l.lockKeyPrefix, key, l.lockSession),
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

	l.key = key

	return nil
}

// Acquire acquires a lock on a key. A time-to-live time is set
// on the lock to ensure the lock is invalidated after the time is passed.
func (l *ConsulLock) Acquire(key string) error {
	var err error

	// If lock object has got a session, get one.
	if l.lockSession == "" {
		l.lockSession, err = l.createSession(int(LockTTL.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to get lock: %s", err)
		}
	}

	return l.acquire(key)
}

// Release invalidates the lock previously acquired
func (l *ConsulLock) Release() error {
	resp, err := goreq.Request{
		Method: "PUT",
		Uri:    l.consulAddr + "/v1/kv/" + fmt.Sprintf("%s.%s?release=%s", l.lockKeyPrefix, l.key, l.lockSession),
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
func (l *ConsulLock) IsAcquirer() error {
	fmt.Println(l.lockSession)
	if len(l.key) == 0 {
		return fmt.Errorf("key is not set")
	}

	resp, err := goreq.Request{
		Uri: l.consulAddr + "/v1/kv/" + fmt.Sprintf("%s.%s", l.lockKeyPrefix, l.key),
	}.Do()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("Not 200", resp.StatusCode)
		return types.ErrLockNotAcquired
	}

	var sessions []map[string]interface{}
	err = resp.Body.FromJsonTo(&sessions)
	if err != nil {
		return err
	}
	util.Printify(sessions)

	if len(sessions) > 0 {
		if sessions[0]["Session"] == nil {
			return types.ErrLockNotAcquired
		}
	} else {
		return types.ErrLockNotAcquired
	}

	return nil
}
