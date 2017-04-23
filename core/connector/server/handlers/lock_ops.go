package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
)

// LockOperations defines a structure for handling lock operations from the cocoon code
type LockOperations struct {
	log *logging.Logger
}

// NewLockOperationHandler creates a new LockOperations instance
func NewLockOperationHandler(log *logging.Logger) *LockOperations {
	return &LockOperations{
		log: log,
	}
}

// Handle handles cocoon operations
func (l *LockOperations) Handle(ctx context.Context, op *proto_connector.LockOperation) (*proto_connector.Response, error) {
	switch op.Name {
	case types.OpLockAcquire:
		return l.acquire(op)
	case types.OpLockCheckAcquire:
		return l.isAcquirer(op)
	case types.OpLockRelease:
		return l.release(op)
	default:
		return nil, fmt.Errorf("unsupported lock operation")
	}
}

// acquire a lock
func (l *LockOperations) acquire(op *proto_connector.LockOperation) (*proto_connector.Response, error) {
	if len(op.Params) != 4 {
		return nil, fmt.Errorf("unexpected number of parameters. Expects 4 parameters")
	}

	if len(op.Params[0]) == 0 {
		return nil, fmt.Errorf("cocoon id is required")
	}

	if len(op.Params[1]) == 0 {
		return nil, fmt.Errorf("key is required")
	}

	if len(op.Params[2]) == 0 {
		return nil, fmt.Errorf("ttl is required")
	}

	ttl, _ := strconv.Atoi(op.Params[2])
	ttlDur := time.Duration(ttl) * time.Second

	// create lock and set default state
	lock := common.NewLockWithTTL(op.Params[1], ttlDur)
	lock.SetState(map[string]interface{}{
		"lock_key_prefix": fmt.Sprintf("platform/lock/%s", op.Params[0]),
		"lock_session":    op.Params[3],
	})

	if err := lock.Acquire(); err != nil {
		return nil, err
	}

	lockSession := lock.GetState()["lock_session"].(string)

	return &proto_connector.Response{
		Body:   []byte(lockSession),
		Status: 200,
	}, nil
}

// isAcquirer checks whether a lock is acquired by a lock session
func (l *LockOperations) isAcquirer(op *proto_connector.LockOperation) (*proto_connector.Response, error) {
	if len(op.Params) != 3 {
		return nil, fmt.Errorf("unexpected number of parameters. Expects 3 parameters")
	}

	if len(op.Params[0]) == 0 {
		return nil, fmt.Errorf("cocoon id is required")
	}

	if len(op.Params[1]) == 0 {
		return nil, fmt.Errorf("key is required")
	}

	if len(op.Params[2]) == 0 {
		return nil, fmt.Errorf("lock session is required")
	}

	// create lock and set default state
	lock := common.NewLock(op.Params[1])
	lock.SetState(map[string]interface{}{
		"lock_key_prefix": fmt.Sprintf("platform/lock/%s", op.Params[0]),
		"lock_session":    op.Params[2],
	})

	if err := lock.IsAcquirer(); err != nil {
		return nil, err
	}

	return &proto_connector.Response{
		Body:   []byte("true"),
		Status: 200,
	}, nil
}

// isAcquirer checks whether a lock is acquired by a lock session
func (l *LockOperations) release(op *proto_connector.LockOperation) (*proto_connector.Response, error) {
	if len(op.Params) != 3 {
		return nil, fmt.Errorf("unexpected number of parameters. Expects 3 parameters")
	}

	if len(op.Params[0]) == 0 {
		return nil, fmt.Errorf("cocoon id is required")
	}

	if len(op.Params[1]) == 0 {
		return nil, fmt.Errorf("key is required")
	}

	if len(op.Params[2]) == 0 {
		return nil, fmt.Errorf("lock session is required")
	}

	// create lock and set default state
	lock := common.NewLock(op.Params[1])
	lock.SetState(map[string]interface{}{
		"lock_key_prefix": fmt.Sprintf("platform/lock/%s", op.Params[0]),
		"lock_session":    op.Params[2],
	})

	if err := lock.Release(); err != nil {
		return nil, err
	}

	return &proto_connector.Response{
		Body:   []byte("true"),
		Status: 200,
	}, nil
}
