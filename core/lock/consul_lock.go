package lock

import (
	"fmt"
	"net/url"
	"time"

	"github.com/franela/goreq"
)

// LockTTL defines max time to live of a lock.
var LockTTL = time.Duration(20 * time.Second)

func init() {
	goreq.SetConnectTimeout(time.Second * 10)
}

// ConsulLock provides lock functionalities based on consul sessions.It implements
// The Lock interface.
type ConsulLock struct {
	consulAddr    string
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
	item := url.Values{}
	item.Set("TTL", ttlStr)
	item.Set("Behaviour", "delete")
	resp, err := goreq.Request{
		Method:      "PUT",
		Uri:         l.consulAddr + "/v1/session/create",
		QueryString: item,
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
		Uri:    l.consulAddr + "/v1/kv/" + fmt.Sprintf("%s?acquire=%s", l.lockKeyPrefix, l.lockSession),
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
	fmt.Println("\n", status)
	if status == "false" {
		return fmt.Errorf("lock already acquired by another process")
	}

	return nil
}

// Acquire acquires a lock on a key. A time-to-live time is set
// on the lock to ensure the lock is invalidated after the time is passed.
func (l *ConsulLock) Acquire(key string) error {

	var err error

	l.lockSession, err = l.createSession(int(LockTTL.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to get lock: %s", err)
	}

	err = l.acquire(key)

	return err
}

// Release invalidates the lock previously acquired
func (l *ConsulLock) Release() error {
	return nil
}
