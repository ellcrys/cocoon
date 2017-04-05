package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	context "golang.org/x/net/context"
)

// Deploy starts a new cocoon. The scheduler creates a job based on the requests
func (api *API) Deploy(ctx context.Context, req *proto.DeployRequest) (*proto.Response, error) {

	var err error

	if _, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	resp, err := api.GetCocoon(ctx, &proto.GetCocoonRequest{
		ID: req.GetCocoonID(),
	})
	if err != nil {
		return nil, err
	}

	var cocoon types.Cocoon
	if err = util.FromJSON(resp.Body, &cocoon); err != nil {
		return nil, common.JSONCoerceErr("cocoon", err)
	}

	// Since the number of signatories is greater than one, we need to check if the
	// release has been approved by a minimum of the required threshold.
	if cocoon.NumSignatories > 1 {

		// get the latest release
		resp, err := api.GetRelease(ctx, &proto.GetReleaseRequest{
			ID: cocoon.Releases[len(cocoon.Releases)-1],
		})
		if err != nil && err != types.ErrTxNotFound {
			return nil, fmt.Errorf("failed to get release. %s", err)
		} else if err == types.ErrTxNotFound {
			return nil, fmt.Errorf("failed to get release")
		}

		var release types.Release
		if err = util.FromJSON(resp.Body, &release); err != nil {
			return nil, common.JSONCoerceErr("release", err)
		}

		if release.SigApproved < cocoon.SigThreshold {
			return nil, fmt.Errorf("deployment denied. You currently have %d approval vote(s) of the required %d vote(s)", release.SigApproved, cocoon.SigThreshold)
		}
	}

	depInfo, err := api.scheduler.Deploy(
		req.GetCocoonID(),
		req.GetLanguage(),
		req.GetURL(),
		req.GetReleaseTag(),
		string(req.GetBuildParam()),
		req.GetLink(),
		req.GetMemory(),
		req.GetCPUShares(),
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "system") {
			apiLog.Error(err.Error())
			return nil, fmt.Errorf("failed to deploy cocoon")
		}
		return nil, err
	}

	err = api.updateCocoonStatusOnStarted(ctx, &cocoon)
	if err != nil {
		return nil, fmt.Errorf("failed to update status")
	}

	apiLog.Infof("Successfully deployed cocoon code %s", depInfo.ID)

	return &proto.Response{
		Status: 200,
		Body:   []byte(depInfo.ID),
	}, nil
}

// updateCocoon creates a new cocoon by overriding an existing one (if any).
func (api *API) updateCocoon(ctx context.Context, cocoon *types.Cocoon) error {
	var protoCreateCocoonReq proto.CreateCocoonRequest
	if err := cstructs.Copy(cocoon, &protoCreateCocoonReq); err != nil {
		return err
	}
	protoCreateCocoonReq.OptionAllowDuplicate = true
	_, err := api.CreateCocoon(ctx, &protoCreateCocoonReq)
	if err != nil {
		return err
	}
	return nil
}

// updateCocoonStatusOnStarted checks on interval the cocoon status and update the
// status field when the cocoon status is `started`. It returns immediately
// an error is encountered. The function will block till success or failure.
func (api *API) updateCocoonStatusOnStarted(ctx context.Context, cocoon *types.Cocoon) error {
	for {
		status, err := api.GetCocoonStatus(cocoon.ID)
		if err != nil {
			return err
		}
		if status == CocoonStatusRunning {
			cocoon.Status = CocoonStatusStarted
			return api.updateCocoon(ctx, cocoon)
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}
