package api

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	context "golang.org/x/net/context"
)

// CreateRelease creates a release
func (api *API) CreateRelease(ctx context.Context, req *proto_api.CreateReleaseRequest) (*proto_api.Response, error) {

	var release types.Release
	cstructs.Copy(req, &release)
	release.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	// recreate ACLMap from byte value in request
	if len(req.ACL) > 0 {
		var aclMap map[string]interface{}
		if err := util.FromJSON(req.ACL, &aclMap); err != nil {
			return nil, fmt.Errorf("acl: malformed json")
		}
		release.ACL = types.NewACLMap(aclMap)
	}

	req = nil

	if err := ValidateRelease(&release); err != nil {
		return nil, err
	}

	_, err := api.platform.GetRelease(ctx, release.ID, false)
	if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, err
	} else if err == nil {
		return nil, fmt.Errorf("a release with matching id already exists")
	}

	err = api.platform.PutRelease(ctx, &release)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   release.ToJSON(),
	}, nil
}

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
