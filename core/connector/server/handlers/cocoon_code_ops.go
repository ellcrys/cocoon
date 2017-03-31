package handlers

import (
	"context"
	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/connector_proto"
	"github.com/ncodes/cocoon/core/runtime/golang/proto"
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
func (l *CocoonCodeOperations) Handle(ctx context.Context, op *connector_proto.CocoonCodeOperation) (*connector_proto.Response, error) {

	client, err := grpc.Dial(l.cocoonCodeRPCAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer client.Close()

	stub := proto.NewStubClient(client)
	resp, err := stub.Invoke(ctx, &proto.InvokeParam{
		ID:       util.UUID4(),
		Function: op.GetFunction(),
		Params:   op.GetParams(),
	})

	if err != nil {
		return nil, fmt.Errorf("invoke error: %s", common.GetRPCErrDesc(err))
	}

	return &connector_proto.Response{
		ID:     resp.ID,
		Status: 200,
		Body:   resp.Body,
	}, nil
}
