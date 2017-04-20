package api

import (
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

// Deploy instructs the scheduler to start a cocoon. The latest release is
// fetched and validated to ensure it has enough votes. If the required votes
// are available, the cocoon is updated with the release value and also executed.
func (api *API) Deploy(ctx context.Context, req *proto_api.DeployRequest) (*proto_api.Response, error) {

	apiLog.Infof("New deploy request for cocoon = [%s]", req.CocoonID)

	var err error
	var claims jwt.MapClaims

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	resp, err := api.GetCocoon(ctx, &proto_api.GetCocoonRequest{ID: req.GetCocoonID()})
	if err != nil {
		return nil, err
	}

	var cocoon types.Cocoon
	util.FromJSON(resp.Body, &cocoon)
	userSessionIdentity := claims["identity"].(string)

	// ensure logged in user owns this cocoon
	if userSessionIdentity != cocoon.IdentityID {
		return nil, fmt.Errorf("Permission denied: You do not have permission to perform this operation")
	}

	// don't continue if cocoon is started or running
	if cocoon.Status == CocoonStatusStarted || cocoon.Status == CocoonStatusRunning {
		if cocoon.Status == CocoonStatusStarted {
			return nil, fmt.Errorf("cocoon has already been started")
		}
		return nil, fmt.Errorf("cocoon is already running")
	}

	// don't continue if cocoon has no release (this should never happen)
	if len(cocoon.Releases) == 0 {
		return nil, fmt.Errorf("No release to run")
	}

	protoReleaseReq := &proto_api.GetReleaseRequest{
		ID: cocoon.Releases[len(cocoon.Releases)-1],
	}

	// if user wants to use the last deployed release, set the release request id to the last deployed release id
	// otherwise, if the cocoon does not have a last deployed release, return error
	if req.UseLastDeployedRelease && len(cocoon.LastDeployedRelease) != 0 {
		apiLog.Infof("Using last deployed release for cocoon = [%s]", cocoon.ID)
		protoReleaseReq.ID = cocoon.LastDeployedRelease
	} else if req.UseLastDeployedRelease {
		return nil, fmt.Errorf("this cocoon does not have a recently approved and deployed release yet")
	}

	// get the release
	resp, err = api.GetRelease(ctx, protoReleaseReq)
	if err != nil && err != types.ErrTxNotFound {
		return nil, fmt.Errorf("failed to get release. %s", err)
	} else if err == types.ErrTxNotFound {
		return nil, fmt.Errorf("release (%s) does not exist", protoReleaseReq.ID)
	}

	var release types.Release
	util.FromJSON(resp.Body, &release)

	// If the number of signatories is greater than 1 and the number of approval
	// signatories for the release is less than the set signatory threshold, we cannot start this cocoon
	if cocoon.NumSignatories > 1 && release.SigApproved < cocoon.SigThreshold {
		return nil, fmt.Errorf(
			"denied. This cocoon has not met the required number of approval votes.\nRelease ID (Latest): %s\nRequired Number of Approval Votes: %d\nApproval Votes Received: %d\nDeny Votes Received: %d ",
			release.ID,
			cocoon.SigThreshold,
			release.SigApproved,
			release.SigDenied,
		)
	}

	// update the cocoon values to match the release we are about to start
	cocoon.Language = release.Language
	cocoon.URL = release.URL
	cocoon.Version = release.Version
	cocoon.BuildParam = string(release.BuildParam)
	cocoon.Link = release.Link
	cocoon.LastDeployedRelease = release.ID
	cocoon.Firewall = release.Firewall
	cocoon.ACL = release.ACL

	depInfo, err := api.scheduler.Deploy(cocoon.ID, cocoon.Language, cocoon.URL, cocoon.Version, cocoon.BuildParam, cocoon.Link, cocoon.Memory, cocoon.CPUShare)
	if err != nil {
		if strings.HasPrefix(err.Error(), "system") {
			apiLog.Error(err.Error())
			return nil, fmt.Errorf("failed to deploy cocoon")
		}
		return nil, err
	}

	// watch the cocoon status and react accordingly
	err = api.watchCocoonStatus(ctx, &cocoon, func(status string, done func()) error {

		// cocoon is running, update status
		if status == CocoonStatusRunning {
			cocoon.Status = CocoonStatusStarted
			err := api.putCocoon(ctx, &cocoon)
			if err != nil {
				return err
			}
			done()
		}

		// cocoon is dead. We need to set the status to 'dead' and ask the scheduler
		// to delete (stop it - according to nomad).
		if status == CocoonStatusDead {
			apiLog.Debugf("Cocoon [%s] is dead", cocoon.ID)
			_, err = api.StopCocoon(ctx, &proto_api.StopCocoonRequest{ID: cocoon.ID})
			if err != nil {
				return fmt.Errorf("failed to stop dead cocoon: %s", err)
			}
			apiLog.Debugf("Stopped dead cocoon [%s]", cocoon.ID)
			cocoon.Status = CocoonStatusDead
			api.putCocoon(ctx, &cocoon)
			return fmt.Errorf("cocoon is dead")
		}
		return nil
	})
	if err != nil {
		if err.Error() == "cocoon is dead" {
			return nil, fmt.Errorf("cocoon did not start. Please try again")
		}
		return nil, err
	}

	apiLog.Infof("Successfully deployed cocoon code %s", depInfo.ID)

	return &proto_api.Response{
		Status: 200,
		Body:   []byte(depInfo.ID),
	}, nil
}
