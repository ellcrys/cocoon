package handlers

import (
	"fmt"

	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/connector/server/proto_connector"
	"github.com/ellcrys/cocoon/core/stub/proto_runtime"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// CocoonCodeOperations represents a cocoon code operation handlers
type CocoonCodeOperations struct {
	cocoonCodeRPCAddr string
}

// NewCocoonCodeHandler creates a new instance of a ledger operation handler
func NewCocoonCodeHandler(cocoonCodeRPCAddr string) *CocoonCodeOperations {
	return &CocoonCodeOperations{
		cocoonCodeRPCAddr: cocoonCodeRPCAddr,
	}
}

// Handle handles cocoon operations
func (l *CocoonCodeOperations) Handle(ctx context.Context, op *proto_connector.CocoonCodeOperation) (*proto_connector.Response, error) {

	client, err := grpc.Dial(l.cocoonCodeRPCAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer client.Close()

	stub := proto_runtime.NewStubClient(client)
	resp, err := stub.Invoke(ctx, &proto_runtime.InvokeParam{
		ID:       op.ID,
		Function: op.GetFunction(),
		Params:   op.GetParams(),
		Header:   op.GetHeader(),
	})

	if err != nil {
		return nil, fmt.Errorf(common.GetRPCErrDesc(err))
	}

	return &proto_connector.Response{
		ID:     resp.ID,
		Status: 200,
		Body:   resp.Body,
	}, nil
}

// Stop the cocoon code
func (l *CocoonCodeOperations) Stop(ctx context.Context) error {
	client, err := grpc.Dial(l.cocoonCodeRPCAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer client.Close()

	stub := proto_runtime.NewStubClient(client)
	_, err = stub.Stop(ctx, new(proto_runtime.Void))
	return err
}
