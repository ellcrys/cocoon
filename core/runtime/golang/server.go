package golang

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/go-errors/errors"
	"github.com/ncodes/cocoon/core/runtime/golang/proto_runtime"
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

// Invoke invokes a function on the running cocoon code
func (server *stubServer) Invoke(ctx context.Context, params *proto_runtime.InvokeParam) (*proto_runtime.InvokeResponse, error) {

	var err error
	var resp = &proto_runtime.InvokeResponse{
		ID: params.GetID(),
	}

	// This closure allows us to catch panics from the cocoon code Invoke() method
	// so cocoon codes will always continue to run
	func() {

		defer func() {
			if rErr := recover(); rErr != nil {
				err = errors.WrapPrefix(err, "Invoke() panicked", 2)
				log.Errorf(err.Error())
			}
		}()

		var result interface{}
		var header = Metadata(map[string]string{"id": params.GetID()})
		result, err = ccode.OnInvoke(header, params.GetFunction(), params.GetParams())
		if err != nil {
			return
		}

		// coerce result to json
		resultJSON, err := util.ToJSON(result)
		if err != nil {
			err = fmt.Errorf("failed to coerce cocoon code Invoke() result to json string. %s", err)
			return
		}

		resp.Status = 200
		resp.Body = resultJSON
	}()

	if err != nil {
		return nil, err
	}

	return resp, nil
}
