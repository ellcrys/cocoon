package api

import (
	context "golang.org/x/net/context"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
)

// GetLogs fetches logs
func (api *API) GetLogs(ctx context.Context, req *proto.GetLogsRequest) (*proto.Response, error) {

	messages, err := api.logProvider.Get(ctx, fmt.Sprintf("connector-7af6bfd7-eac5-8cb6-3371-8f6608f1f093"), int(req.NumLines), req.Source)
	if err != nil {
		return nil, err
	}

	messagesBytes, _ := util.ToJSON(messages)

	return &proto.Response{
		Status: 200,
		Body:   messagesBytes,
	}, nil
}
