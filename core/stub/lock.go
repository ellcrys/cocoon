package stub

import (
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	"github.com/ncodes/cocoon/core/types"
)

// ErrLockNotAcquired defines an error about a lock not acquired
var ErrLockNotAcquired = fmt.Errorf("lock not acquired")

// Lock defines a structure for a platform-wide lock mechanism
type Lock struct {
	cocoonID      string
	key           string
	lockSessionID string
	ttl           int
}

// newLock creates a lock instance
func newLock(cocoonID, key string, ttl time.Duration) (*Lock, error) {

	key = strings.TrimSpace(key)
	if len(key) == 0 {
		return nil, fmt.Errorf("key is not set")
	}

	// ensure ttl is not lower than 10 seconds and greater than 30 minutes
	ttlSec := int(ttl.Seconds())
	if ttlSec < 10 {
		ttlSec = 10
	}

	if ttlSec > (60 * 30) {
		return nil, fmt.Errorf("ttl cannot be greater than 30 minutes (3600 seconds)")
	}

	return &Lock{
		cocoonID: cocoonID,
		key:      key,
		ttl:      ttlSec,
	}, nil
}

// Acquire will attempt to acquire a lock on the key within the
// cocoon scope defined in the lock instance. If successful, Lock object is returned.
// An acquired lock will also be enforced in other linked cocoon codes.
func (l *Lock) Acquire() error {
	lockSessionID, err := sendLockOp(&proto_connector.LockOperation{
		Name:   types.OpLockAcquire,
		Params: []string{l.cocoonID, l.key, strconv.Itoa(l.ttl), l.lockSessionID},
		LinkTo: l.cocoonID,
	})
	if err != nil {
		if common.CompareErr(err, ErrLockNotAcquired) == 0 {
			return ErrLockNotAcquired
		}
		return err
	}
	l.lockSessionID = string(lockSessionID)
	return nil
}

// IsAcquirer checks whether lock session is still active.
// If the lock is still active, nil is returned.
func (l *Lock) IsAcquirer() error {

	if len(l.lockSessionID) == 0 {
		return fmt.Errorf("lock session not set")
	}

	_, err := sendLockOp(&proto_connector.LockOperation{
		Name:   types.OpLockCheckAcquire,
		Params: []string{l.cocoonID, l.key, l.lockSessionID},
		LinkTo: l.cocoonID,
	})
	if err != nil {
		if common.CompareErr(err, ErrLockNotAcquired) == 0 {
			return ErrLockNotAcquired
		}
		return err
	}

	return nil
}

// Release release the lock. Returns
// nil if successful or if no lock was held.
func (l *Lock) Release() error {

	if len(l.lockSessionID) == 0 {
		return nil
	}

	_, err := sendLockOp(&proto_connector.LockOperation{
		Name:   types.OpLockRelease,
		Params: []string{l.cocoonID, l.key, l.lockSessionID},
		LinkTo: l.cocoonID,
	})
	if err != nil {
		if common.CompareErr(err, ErrLockNotAcquired) == 0 {
			return ErrLockNotAcquired
		}
		return err
	}

	return nil
}
