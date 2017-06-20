package stub

import (
	"fmt"

	"github.com/ellcrys/cocoon/core/stub/proto_runtime"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// StubServer defines the services of the stub's GRPC connection
type stubServer struct {
	port int
}

// HealthCheck returns health status
func (server *stubServer) HealthCheck(context.Context, *proto_runtime.Ok) (*proto_runtime.Ok, error) {
	return &proto_runtime.Ok{
		Status: 200,
	}, nil
}

func recoverPanic(f func()) error {
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case error:
					err = errors.Wrap(v, "Panicked")
				default:
					err = errors.Wrap(fmt.Errorf("%v", v), "Panicked")
				}
				log.Errorf("%+v", err)
			}
		}()
		f()
	}()

	return err
}

// Invoke invokes a function on the running cocoon code
func (server *stubServer) Invoke(ctx context.Context, params *proto_runtime.InvokeParam) (*proto_runtime.InvokeResponse, error) {

	var err error
	var resp = &proto_runtime.InvokeResponse{
		ID: params.GetID(),
	}

	// This closure allows us to catch panics from the cocoon code Invoke() method
	// so cocoon codes will always continue to run
	err = recoverPanic(func() {
		result, err := ccode.OnInvoke(params.GetHeader(), params.GetFunction(), params.GetParams())
		if err != nil {
			return
		}
		resp.Status = 200
		resp.Body = result
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Stop calls the stop method of the code
func (server *stubServer) Stop(context.Context, *proto_runtime.Void) (*proto_runtime.Void, error) {
	ccode.OnStop()
	return new(proto_runtime.Void), nil
}
