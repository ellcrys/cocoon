package connector

import (
	"time"

	"github.com/ncodes/cocoon/core/runtime/golang/proto_runtime"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// HealthChecker checks the health status of
// the cocoon code. It repeatedly calls the cocoon coder
// health check method. If cocoon code refuses to respond,
// it calls the OnDeadFunc method attached to it
type HealthChecker struct {
	OnDeadFunc     func()
	cocoonCodeAddr string
	ticker         *time.Ticker
}

// NewHealthChecker creates a cocoon code health checker instance
func NewHealthChecker(cocoonCodeAddr string, onDeadFunc func()) *HealthChecker {
	return &HealthChecker{
		cocoonCodeAddr: cocoonCodeAddr,
		OnDeadFunc:     onDeadFunc,
	}
}

// Start runs the health check immediately and on future intervals.
// if check returns err, it calls the OnDeadFunc and stops the health check.
func (hc *HealthChecker) Start() {

	logHealthChecker.Infof("Started health check on cocoon code @ %s", hc.cocoonCodeAddr)

	if err := hc.check(); err != nil {
		logHealthChecker.Infof(err.Error())
		if hc.OnDeadFunc != nil {
			hc.OnDeadFunc()
		}
		return
	}

	hc.ticker = time.NewTicker(15 * time.Second)
	for _ = range hc.ticker.C {
		if hc.check() != nil {
			if hc.OnDeadFunc != nil {
				hc.OnDeadFunc()
			}
			hc.Stop()
		}
	}
}

// check calls the check endpont of the stub.
// It will retry the health check for upto 5 times with
// a 1 second wait duration if health check fails.
func (hc *HealthChecker) check() error {
	retryLimit := 5
	for retryLimit > 0 {
		client, err := grpc.Dial(hc.cocoonCodeAddr, grpc.WithInsecure())
		if err != nil {
			return err
		}

		stub := proto_runtime.NewStubClient(client)
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = stub.HealthCheck(ctx, &proto_runtime.Ok{})
		if err != nil {
			retryLimit--
			time.Sleep(1 * time.Second)
			return err
		}
		return nil
	}
	return nil
}

// Stop health check
func (hc *HealthChecker) Stop() {
	if hc.ticker != nil {
		hc.ticker.Stop()
	}
}
