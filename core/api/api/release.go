package api

import (
	"github.com/ellcrys/cocoon/core/api/api/proto_api"
	"github.com/ellcrys/cocoon/core/types"
	context "golang.org/x/net/context"
)

// GetRelease returns a release
func (api *API) GetRelease(ctx context.Context, req *proto_api.GetReleaseRequest) (*proto_api.Response, error) {

	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)

	release, err := api.platform.GetRelease(ctx, req.ID, false)
	if err != nil {
		return nil, err
	}

	cocoon, err := api.platform.GetCocoon(ctx, release.CocoonID)
	if err != nil {
		return nil, err
	}

	// if logged in user owns the release's cocoon, refetch the cocoon and include private fields
	if loggedInIdentity == cocoon.IdentityID {
		release, err = api.platform.GetRelease(ctx, req.ID, true)
		if err != nil {
			return nil, err
		}
	}

	return &proto_api.Response{
		Body:   []byte(release.ToJSON()),
		Status: 200,
	}, nil
}
