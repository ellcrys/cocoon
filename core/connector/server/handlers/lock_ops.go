package handlers

import (
	"fmt"
	"strconv"
	"time"

	"strings"

	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/connector"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
)

// LockOperations defines a structure for handling lock operations from the cocoon code
type LockOperations struct {
	log       *logging.Logger
	connector *connector.Connector
}

// NewLockOperationHandler creates a new LockOperations instance
func NewLockOperationHandler(log *logging.Logger, connector *connector.Connector) *LockOperations {
	return &LockOperations{
		log:       log,
		connector: connector,
	}
}

// checkPermission checks whether the request has permission to permform lock operations
// against a target cocoon. Current permission rules are:
// 1. A lock request is allowed if the target cocoon is a system cocoon and the key does
// not begin with an underscore (keys begining with an underscore are reservered as system specific keys).
// 2. A lock request is allowed if the requesting cocoon is natively linked to the target cocoon.
func (l *LockOperations) checkPermission(ctx context.Context, op *proto_connector.LockOperation) error {

	// Target cocoon is the system cocoon
	if op.LinkTo == types.SystemCocoonID {
		if strings.HasPrefix(op.Params[1], "_") {
			return fmt.Errorf("reserved key error: cannot create/access a lock with key starting with an underscore on system cocoon")
		}
		return nil
	}

	// Target cocoon is not the same as the current cocoon
	if op.LinkTo != l.connector.GetRequest().ID {

		// get the currently executing release of the current cocoon
		_, release, err := l.connector.Platform.GetCocoonAndRelease(ctx, l.connector.GetRequest().ID, l.connector.GetRequest().ReleaseID, false)
		if err != nil {
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				return fmt.Errorf("calling cocoon not found")
			}
			return common.SimpleGRPCError("orderer", err)
		}

		// If current cocoon is not natively linked to the target cocoon, disallow with an error
		if release.Link != op.LinkTo {
			return fmt.Errorf("permission denied: current cocoon not natively linked to target cocoon")
		}
	}
	return nil
}

// Handle handles cocoon operations
func (l *LockOperations) Handle(ctx context.Context, op *proto_connector.LockOperation) (*proto_connector.Response, error) {

	if err := l.checkPermission(ctx, op); err != nil {
		return nil, err
	}

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
	lock, err := common.NewLockWithTTL(op.Params[1], ttlDur)
	if err != nil {
		return nil, err
	}
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
	lock, err := common.NewLock(op.Params[1])
	if err != nil {
		return nil, err
	}
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
	lock, err := common.NewLock(op.Params[1])
	if err != nil {
		return nil, err
	}
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
