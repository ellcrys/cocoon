package api

import (
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

// Deploy starts a new cocoon. The scheduler creates a job based on the requests
func (api *API) Deploy(ctx context.Context, req *proto.DeployRequest) (*proto.Response, error) {

	var err error
	var claims jwt.MapClaims
	var identity string

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	identity = claims["identity"].(string)

	_, err = api.GetCocoon(ctx, &proto.GetCocoonRequest{
		ID:       req.GetID(),
		Identity: identity,
	})
	if err != nil {
		return nil, err
	}

	depInfo, err := api.scheduler.Deploy(
		req.GetID(),
		req.GetLanguage(),
		req.GetURL(),
		req.GetReleaseTag(),
		string(req.GetBuildParam()),
		req.GetLink(),
		req.GetMemory(),
		req.GetCPUShare(),
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "system") {
			apiLog.Error(err.Error())
			return nil, fmt.Errorf("failed to deploy cocoon")
		}
		return nil, err
	}

	apiLog.Infof("Successfully deployed cocoon code %s", depInfo.ID)

	return &proto.Response{
		ID:     req.GetID(),
		Status: 200,
		Body:   []byte(depInfo.ID),
	}, nil
}
