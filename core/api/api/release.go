package api

import (
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	context "golang.org/x/net/context"
)

// GetRelease returns a release
func (api *API) GetRelease(ctx context.Context, req *proto_api.GetReleaseRequest) (*proto_api.Response, error) {

	release, err := api.platform.GetRelease(ctx, req.ID, false)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Body:   []byte(release.ToJSON()),
		Status: 200,
	}, nil
}
