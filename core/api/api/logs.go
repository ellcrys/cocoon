package api

import (
	context "golang.org/x/net/context"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
)

// GetLogs fetches logs
func (api *API) GetLogs(ctx context.Context, req *proto_api.GetLogsRequest) (*proto_api.Response, error) {

	var err error

	_, err = api.platform.GetCocoon(ctx, req.CocoonID)
	if err != nil {
		return nil, err
	}

	// cap lines to fetch at 5000
	if req.NumLines > 5000 {
		req.NumLines = 5000
	}

	messages, err := api.logProvider.Get(ctx, fmt.Sprintf("cocoon-%s", req.CocoonID), int(req.NumLines), req.Source)
	if err != nil {
		if common.CompareErr(err, fmt.Errorf("Invalid resource: id is empty")) == 0 {
			return nil, fmt.Errorf("failed to get logs for cocoon (%s)", req.CocoonID)
		}
		return nil, err
	}

	messagesBytes, _ := util.ToJSON(messages)

	return &proto_api.Response{
		Status: 200,
		Body:   messagesBytes,
	}, nil
}
