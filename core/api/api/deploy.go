package api

import (
	"fmt"
	"strings"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

// Deploy instructs the scheduler to start a cocoon. The latest release is
// fetched and validated to ensure it has enough votes. If the required votes
// are available, the cocoon is updated with the release value and also executed.
func (api *API) Deploy(ctx context.Context, req *proto_api.DeployRequest) (*proto_api.Response, error) {

	var err error
	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)

	apiLog.Infof("New deploy request for cocoon with ID = %s", req.CocoonID)

	cocoon, err := api.platform.GetCocoon(ctx, req.GetCocoonID())
	if err != nil {
		return nil, err
	}

	// ensure logged in user owns this cocoon
	if loggedInIdentity != cocoon.IdentityID {
		return nil, types.ErrPermissionNotGrant
	}

	// don't continue if cocoon's status is in one of these statuses
	statuses := []string{CocoonStatusStarted, CocoonStatusRunning, CocoonStatusBuilding, CocoonStatusDead}
	if util.InStringSlice(statuses, cocoon.Status) {
		switch cocoon.Status {
		case CocoonStatusStarted:
			return nil, fmt.Errorf("cocoon has already been started")
		case CocoonStatusBuilding:
			return nil, fmt.Errorf("cocoon is being built")
		case CocoonStatusRunning:
			return nil, fmt.Errorf("Cocoon is already running")
		case CocoonStatusDead:
			if err := api.stopCocoon(ctx, cocoon.ID); err != nil {
				return nil, fmt.Errorf("failed to stop dead cocoon")
			}
		}
	}

	// when release id isn't provided, use the most recently created release
	if len(req.ReleaseID) == 0 {
		// don't continue if cocoon has no release (this should never happen)
		if len(cocoon.Releases) == 0 {
			return nil, fmt.Errorf("No release to run. Wierd")
		}
		req.ReleaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}

	apiLog.Debugf("Deploying release = %s", req.ReleaseID)

	// get the release
	release, err := api.platform.GetRelease(ctx, req.ReleaseID, true)
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, fmt.Errorf("release (%s) does not exist", req.ReleaseID)
		}
		return nil, fmt.Errorf("failed to get release. %s", err)
	}

	// If the number of signatories is greater than 1 and the number of approval
	// signatories for the release is less than the set signatory threshold, we cannot start this cocoon
	if cocoon.NumSignatories > 1 && release.SigApproved < cocoon.SigThreshold {
		return nil, fmt.Errorf(
			"denied. This cocoon has not met the required number of approval votes.\nRelease ID: %s\nRequired Number of Approval Votes: %d\nApproval Votes Received: %d\nDeny Votes Received: %d ",
			release.ID,
			cocoon.SigThreshold,
			release.SigApproved,
			release.SigDenied,
		)
	}

	// update the cocoon values to match the release we are about to start
	cocoon.LastDeployedReleaseID = release.ID

	depInfo, err := api.platform.GetScheduler().Deploy(cocoon.ID, release.ID, cocoon.Memory, cocoon.CPUShare)
	if err != nil {
		if strings.HasPrefix(err.Error(), "system") {
			apiLog.Error(err.Error())
			return nil, fmt.Errorf("failed to deploy cocoon")
		}
		return nil, err
	}

	// watch the cocoon status and react accordingly
	err = api.watchCocoonStatus(ctx, cocoon, func(status string, done func()) error {

		// cocoon is started, update status
		if status == CocoonStatusRunning {
			cocoon.Status = CocoonStatusRunning
			err := api.platform.PutCocoon(ctx, cocoon)
			if err != nil {
				return err
			}
			done()
			return nil
		}

		// cocoon is dead. We need to set the status to 'dead' and ask the scheduler
		// to delete (stop it - according to nomad).
		if status == CocoonStatusDead {
			apiLog.Debugf("Cocoon [%s] is dead", cocoon.ID)
			err := api.stopCocoon(ctx, cocoon.ID)
			if err != nil {
				return fmt.Errorf("failed to stop dead cocoon: %s", err)
			}
			apiLog.Debugf("Stopped dead cocoon [%s]", cocoon.ID)
			cocoon.Status = CocoonStatusDead
			api.platform.PutCocoon(ctx, cocoon)
			return fmt.Errorf("cocoon died :(")
		}

		// Non-pending status should force the watch process to exit
		if status != "pending" {
			apiLog.Debug("Unrecognised Status: ", status)
			done()
			return nil
		}

		// return nil to keep watching.
		return nil
	})
	if err != nil {
		return nil, err
	}

	apiLog.Infof("Successfully deployed cocoon code %s", depInfo.ID)

	return &proto_api.Response{
		Status: 200,
		Body:   []byte(depInfo.ID),
	}, nil
}
