package api

import (
	"fmt"
	"strings"

	"github.com/ncodes/cocoon/core/api/api/proto"
	context "golang.org/x/net/context"
)

// Deploy starts a new cocoon. The scheduler creates a job based on the requests
func (api *API) Deploy(ctx context.Context, req *proto.DeployRequest) (*proto.Response, error) {
	depInfo, err := api.scheduler.Deploy(
		req.GetID(),
		req.GetLanguage(),
		req.GetURL(),
		req.GetReleaseTag(),
		string(req.GetBuildParam()),
		req.GetMemory(),
		req.GetCPUShare(),
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "system") {
			log.Error(err.Error())
			return nil, fmt.Errorf("failed to deploy cocoon")
		}
		return nil, err
	}

	log.Infof("Successfully deployed cocoon code %s", depInfo.ID)

	return &proto.Response{
		ID:     req.GetID(),
		Status: 200,
		Body:   []byte(depInfo.ID),
	}, nil
}
