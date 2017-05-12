package api

import (
	"fmt"
	"time"

	"os"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/util"
	"github.com/fatih/structs"
	"github.com/jinzhu/copier"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/api/archiver"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

var (
	// CocoonStatusCreated indicates a created cocoon
	CocoonStatusCreated = "created"

	// CocoonStatusStarted indicates a started cocoon cocoon
	CocoonStatusStarted = "started"

	// CocoonStatusRunning indicates a running cocoon code
	CocoonStatusRunning = "running"

	// CocoonStatusBuilding indicates a cocoon in build phase
	CocoonStatusBuilding = "building"

	// CocoonStatusStopped indicates a stopped cocoon
	CocoonStatusStopped = "stopped"

	// CocoonStatusDead indicates a dead cocoon
	CocoonStatusDead = "dead"
)

// watchCocoonStatus checks the status of a cocoon on interval and passes it to a callback function.
// The callback is also passed a `done` function to be called to stop the status check. Returning
// error from the callback will also stop the status check.
// This function blocks the current goroutine.
func (api *API) watchCocoonStatus(ctx context.Context, cocoon *types.Cocoon, callback func(s string, doneFunc func()) error) error {
	var done = false
	var err error
	var status string
	for !done {
		status, err = api.GetCocoonStatus(cocoon.ID)
		if err != nil {
			break
		}
		err = callback(status, func() {
			done = true
		})
		if err != nil {
			done = true
			continue
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

// CreateCocoon creates a new cocoon and initial release. The new
// cocoon is also added to the identity's list of cocoons
func (api *API) CreateCocoon(ctx context.Context, req *proto_api.ContractRequest) (*proto_api.Response, error) {

	var err error
	var releaseID = util.UUID4()
	var now = time.Now()
	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)
	var cocoon types.Cocoon

	cocoon.Merge(structs.New(req).Map())
	cocoon.ID = req.CocoonID
	cocoon.Status = CocoonStatusCreated
	cocoon.Releases = []string{releaseID}
	cocoon.IdentityID = loggedInIdentity
	cocoon.Signatories = append(cocoon.Signatories, cocoon.IdentityID)
	cocoon.CreatedAt = now.UTC().Format(time.RFC3339Nano)

	// create new release
	var release types.Release
	copier.Copy(&release, req)
	release.ID = releaseID
	release.CocoonID = cocoon.ID
	release.ACL = types.NewACLMapFromByte(req.ACL)
	release.CreatedAt = now.UTC().Format(time.RFC3339Nano)

	// Resolve firewall rules destination if firewall is enabled.
	// Otherwise, set firewall to nil
	if req.EnableFirewall {
		if len(release.Firewall) > 0 {
			release.Firewall, err = common.ResolveFirewall(release.Firewall.DeDup())
			if err != nil {
				return nil, fmt.Errorf("firewall: %s", err)
			}
		}
	} else {
		release.Firewall = nil
	}

	// ensure a similar cocoon does not exist
	_, err = api.platform.GetCocoon(ctx, cocoon.ID)
	if err != nil && err != types.ErrCocoonNotFound {
		return nil, err
	} else if err == nil {
		return nil, fmt.Errorf("cocoon with the same id already exists")
	}

	if err := ValidateCocoon(&cocoon); err != nil {
		return nil, err
	}

	if err := ValidateRelease(&release); err != nil {
		return nil, err
	}

	// if a link cocoon id is provided, ensure the linked cocoon exists
	// and is owned by the currently logged in identity
	if len(req.Link) > 0 {
		cocoonToLinkTo, err := api.platform.GetCocoon(ctx, req.Link)
		if err != nil {
			if err != types.ErrCocoonNotFound {
				return nil, err
			} else if err == types.ErrCocoonNotFound {
				return nil, fmt.Errorf("link: cannot link to a non-existing cocoon %s", req.Link)
			}
		}
		// ensure logged in user owns the cocoon being linked
		if loggedInIdentity != cocoonToLinkTo.IdentityID {
			return nil, fmt.Errorf("link: Permission denied. Cannot create a native link to a cocoon you did not create")
		}
	}

	err = api.platform.PutCocoon(ctx, &cocoon)
	if err != nil {
		return nil, err
	}

	err = api.platform.PutRelease(ctx, &release)
	if err != nil {
		return nil, err
	}

	// Include new cocoon id in the logged in user's cocoon list
	identity, err := api.platform.GetIdentity(ctx, cocoon.IdentityID)
	if err != nil {
		return nil, err
	}
	identity.Cocoons = append(identity.Cocoons, cocoon.ID)
	err = api.platform.PutIdentity(ctx, identity)
	if err != nil {
		return nil, err
	}

	// archive release
	if env := os.Getenv("ENV"); env == "production" || env == "development" {
		persister, err := archiver.NewGStoragePersister(archiver.MakeArchiveName(cocoon.ID, release.Version))
		if err != nil {
			return nil, fmt.Errorf("failed to create persister")
		}
		err = archiver.NewArchiver(
			archiver.NewGitObject(release.URL, release.Version),
			persister,
		).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to archive release: %s", err)
		}
	}

	return &proto_api.Response{
		Status: 200,
		Body:   cocoon.ToJSON(),
	}, nil
}

// updateReleaseEnv processes flag behaviours of pin, unpin and
// unpin_once and include the changes to the latest Env.
// The following are the expected behaviour of the flags:
//
// "pin": All flags in the last release that have "pin" flag will
// overwrite the corresponding/matching flag in latestEnv. In other words, If a variable
// shares same flag name with a pinned flag in the last release, it is completely overridden.
//
// "unpin": This flag prevents "pin" flag from being enforced. It is the antidote of "pin". If a
// variable includes this flag, it cannot be overridden by a matching, pinned variable in the last release.
//
// "unpin_once": This flag is the same as "unpin" except that final variable will still have a "pin" flag
// which is useful if there is need to persist the new update to future releases.
func updateReleaseEnv(lastReleaseEnv, latestEnv types.Env) {

	// get new Env containing pinned variables
	pinnedVarsFromLastRelease := lastReleaseEnv.GetByFlag("pin")

	for curVar, curVal := range latestEnv {
		// Find a variable from the latest Env that is included in the pinned variables on the last release Env
		if pinnedVar, pinnedVarVal, ok := pinnedVarsFromLastRelease.GetFull(curVar); ok {

			curVarFlags := types.GetFlags(curVar)

			// If the current variable has "unpin_once", this means we can't override it's content
			// but we will have to replace the "unpin_once" with "pin" which will keep the variable pinned
			// to future updates
			if util.InStringSlice(types.GetFlags(curVar), "unpin_once") {
				delete(latestEnv, curVar)
				curVar, _ = types.ReplaceFlag(curVar, "unpin_once", "pin")
				latestEnv[curVar] = curVal
				continue
			}

			// If the current variable does not contain an "unpin" flag, then we have to
			// override it!
			if !util.InStringSlice(curVarFlags, "unpin") {
				delete(latestEnv, curVar)
				latestEnv[pinnedVar] = pinnedVarVal
			}
		}
	}
}

// UpdateCocoon updates a cocoon and optionally creates a new
// release. A new release is created when Release related fields are
// changed.
func (api *API) UpdateCocoon(ctx context.Context, req *proto_api.ContractRequest) (*proto_api.Response, error) {

	var err error
	var cocoonUpdated bool
	var releaseUpdated bool
	var versionUpdated bool
	var now = time.Now()
	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)

	cocoon, err := api.platform.GetCocoon(ctx, req.CocoonID)
	if err != nil {
		return nil, err
	}

	// Ensure the cocoon identity matches the logged in user
	if cocoon.IdentityID != loggedInIdentity {
		return nil, types.ErrPermissionNotGrant
	}

	var cocoonUpd = cocoon.Clone()
	cocoonUpd.Merge(structs.New(req).Map())

	// check if the existing cocoon differ from the updated cocoon
	// if so, apply the new update
	if diffs := cocoon.Difference(cocoonUpd); diffs[0] != nil {
		cocoonUpdated = true
		cocoon = &cocoonUpd
		if err = ValidateCocoon(cocoon); err != nil {
			return nil, err
		}
	}

	// Get the most recent deployed release. if no recent deployed release, get the last created release.
	var recentReleaseID = cocoon.LastDeployedReleaseID
	if len(recentReleaseID) == 0 {
		recentReleaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}

	release, err := api.platform.GetRelease(ctx, recentReleaseID, true)
	if err != nil {
		return nil, err
	}

	releaseUpd := release.Clone()
	releaseUpd.Merge(structs.New(req).Map())
	releaseUpd.Env = req.Env
	releaseUpd.BuildParam = req.BuildParam
	releaseUpd.Link = req.Link
	releaseUpd.ACL = types.NewACLMapFromByte(req.ACL)

	// process special "pin", "unpin" and "unpin_once" environment flags
	updateReleaseEnv(release.Env, releaseUpd.Env)

	// Resolve firewall rules destination if firewall is enabled.
	// Otherwise, set firewall to nil
	if req.EnableFirewall {
		if len(releaseUpd.Firewall) > 0 {
			releaseUpd.Firewall, err = common.ResolveFirewall(releaseUpd.Firewall.DeDup())
			if err != nil {
				return nil, fmt.Errorf("firewall: %s", err)
			}
		}
	} else {
		req.Firewall = nil
	}

	// check if the existing release differ from the updated release
	// if so, apply the new update
	if diffs := release.Difference(releaseUpd); diffs[0] != nil {
		releaseUpdated = true
		versionUpdated = release.Version != releaseUpd.Version
		release = &releaseUpd
		release.CreatedAt = now.UTC().Format(time.RFC3339Nano)
		release.ID = util.UUID4()
		if err = ValidateRelease(release); err != nil {
			return nil, err
		}

		// add new id to cocoon's releases
		cocoon.Releases = append(cocoon.Releases, release.ID)
	}

	var finalResp = map[string]interface{}{
		"newReleaseID":  "",
		"cocoonUpdated": cocoonUpdated,
	}

	// create new release if a field was changed
	if releaseUpdated {
		err = api.platform.PutRelease(ctx, release)
		if err != nil {
			return nil, err
		}

		finalResp["newReleaseID"] = release.ID
	}

	// update cocoon if cocoon was changed or release was updated/created
	if cocoonUpdated || releaseUpdated {
		err = api.platform.PutCocoon(ctx, cocoon)
		if err != nil {
			return nil, err
		}
	}

	// archive release version if it changed
	if env := os.Getenv("ENV"); versionUpdated && (env == "production" || env == "development") {
		persister, err := archiver.NewGStoragePersister(archiver.MakeArchiveName(cocoon.ID, release.Version))
		if err != nil {
			return nil, fmt.Errorf("failed to create persister")
		}
		err = archiver.NewArchiver(
			archiver.NewGitObject(release.URL, release.Version),
			persister,
		).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to archive release: %s", err)
		}
	}

	value, _ := util.ToJSON(&finalResp)

	return &proto_api.Response{
		Status: 200,
		Body:   value,
	}, nil
}

// GetCocoon fetches a cocoon
func (api *API) GetCocoon(ctx context.Context, req *proto_api.GetCocoonRequest) (*proto_api.Response, error) {

	cocoon, err := api.platform.GetCocoon(ctx, req.GetID())
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   cocoon.ToJSON(),
	}, nil
}

// GetCocoonStatus fetches the cocoon status.It queries the scheduler
// discovery service to find out the current service status for the cocoon.
// If the scheduler discovery service does not say the cocoon is running,
// we check with the scheduler to know if the cocoon code was deployed successfully.
func (api *API) GetCocoonStatus(cocoonID string) (string, error) {

	sd, err := api.scheduler.GetServiceDiscoverer()
	if err != nil {
		return "", fmt.Errorf("failed to get service discovery from scheduler: %s", err)
	}

	s, err := sd.GetByID("cocoon", map[string]string{"tag": cocoonID})
	if err != nil {
		apiLog.Errorf("failed to query cocoon service status: %s", err.Error())
		return "", fmt.Errorf("failed to query cocoon service status")
	}

	if len(s) == 0 {

		// check with the scheduler to know status of the cocoon deployment
		dStatus, err := api.scheduler.GetDeploymentStatus(cocoonID)
		if err != nil {
			if err.Error() != "not found" {
				return CocoonStatusStopped, nil
			}
			return "", err
		}

		if dStatus == "dead" {
			return CocoonStatusDead, nil
		}

		return dStatus, nil
	}
	return CocoonStatusRunning, nil
}

// stopCocoon stops a cocoon
func (api *API) stopCocoon(ctx context.Context, id string) error {

	cocoon, err := api.platform.GetCocoon(ctx, id)
	if err != nil {
		return err
	}

	apiLog.Info("Calling scheduler to stop cocoon = ", id)
	err = api.scheduler.Stop(id)
	if err != nil {
		apiLog.Error(err.Error())
		return fmt.Errorf("failed to stop cocoon")
	}

	cocoon.Status = CocoonStatusStopped

	if err = api.platform.PutCocoon(ctx, cocoon); err != nil {
		apiLog.Error(err.Error())
		return fmt.Errorf("failed to update cocoon status")
	}

	return nil
}

// StopCocoon stops a running cocoon and sets its status to `stopped`
func (api *API) StopCocoon(ctx context.Context, req *proto_api.StopCocoonRequest) (*proto_api.Response, error) {

	var err error
	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)

	apiLog.Infof("Received request to stop cocoon = %s", req.ID)

	cocoon, err := api.platform.GetCocoon(ctx, req.GetID())
	if err != nil {
		return nil, err
	}

	// ensure session identity matches cocoon identity
	if loggedInIdentity != cocoon.IdentityID {
		return nil, types.ErrPermissionNotGrant
	}

	if err = api.stopCocoon(ctx, req.GetID()); err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   []byte("done"),
	}, nil
}

// AddSignatories adds one ore more signatories to a cocoon
func (api *API) AddSignatories(ctx context.Context, req *proto_api.AddSignatoriesRequest) (*proto_api.Response, error) {

	var added = []string{}
	var errs = []string{}
	var _id = []string{}
	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)
	var err error

	cocoon, err := api.platform.GetCocoon(ctx, req.CocoonID)
	if err != nil {
		return nil, err
	}

	// ensure session identity matches cocoon identity
	if loggedInIdentity != cocoon.IdentityID {
		return nil, types.ErrPermissionNotGrant
	}

	req.IDs = util.UniqueStringSlice(req.IDs)

	// convert email to ID
	for i, id := range req.IDs {
		_id = append(_id, id)
		if govalidator.IsEmail(id) {
			req.IDs[i] = (&types.Identity{Email: id}).GetID()
		}
	}

	// ensure the number of signatories to add will not exceed the total number of required signatories
	availableSignatorySlots := cocoon.NumSignatories - len(cocoon.Signatories)
	if availableSignatorySlots < len(req.IDs) {
		if availableSignatorySlots == 0 {
			return nil, fmt.Errorf("max signatories already added. You can't add more")
		}
		strPl := "signatures"
		if availableSignatorySlots == 1 {
			strPl = "signatory"
		}
		return nil, fmt.Errorf("maximum required signatories cannot be exceeded. You can only add %d more %s", availableSignatorySlots, strPl)
	}

	for i, id := range req.IDs {

		identity, err := api.platform.GetIdentity(ctx, id)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: identity not found", common.GetShortID(_id[i])))
			continue
		}

		// check if identity is already a signatory
		if util.InStringSlice(cocoon.Signatories, id) {
			errs = append(errs, fmt.Sprintf("%s: identity is already a signatory", common.GetShortID(_id[i])))
			continue
		}

		cocoon.Signatories = append(cocoon.Signatories, identity.GetID())
		added = append(added, identity.GetID())
	}

	if len(added) > 0 {
		err := api.platform.PutCocoon(ctx, cocoon)
		if err != nil {
			return nil, err
		}
	}

	r, _ := util.ToJSON(map[string][]string{
		"added": added,
		"errs":  errs,
	})

	return &proto_api.Response{
		Status: 200,
		Body:   r,
	}, nil
}

// AddVote adds a vote to a release where the logged in user is a signatory
func (api *API) AddVote(ctx context.Context, req *proto_api.AddVoteRequest) (*proto_api.Response, error) {

	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)
	var err error

	release, err := api.platform.GetRelease(ctx, req.ReleaseID, false)
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, fmt.Errorf("release not found")
		}
		return nil, err
	}

	cocoon, err := api.platform.GetCocoon(ctx, release.CocoonID)
	if err != nil {
		return nil, err
	}

	// ensure logged in user is a signatory of this cocoon
	if !util.InStringSlice(cocoon.Signatories, loggedInIdentity) {
		return nil, fmt.Errorf("Permission Denied: You are not a signatory to this cocoon")
	}

	// ensure logged in user has not voted before
	if release.VotersID != nil && util.InStringSlice(release.VotersID, loggedInIdentity) {
		return nil, fmt.Errorf("You have already cast a vote for this release")
	}

	if req.Vote == 1 {
		release.SigApproved++
	}
	if req.Vote == 0 {
		release.SigDenied++
	}

	if release.VotersID == nil {
		release.VotersID = []string{loggedInIdentity}
	} else {
		release.VotersID = append(release.VotersID, loggedInIdentity)
	}

	err = api.platform.PutRelease(ctx, release)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   release.ToJSON(),
	}, nil
}

// RemoveSignatories removes one or more signatories
func (api *API) RemoveSignatories(ctx context.Context, req *proto_api.RemoveSignatoriesRequest) (*proto_api.Response, error) {

	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)
	var err error

	cocoon, err := api.platform.GetCocoon(ctx, req.CocoonID)
	if err != nil {
		return nil, err
	}

	// ensure logged user is owner
	if loggedInIdentity != cocoon.IdentityID {
		return nil, fmt.Errorf("Permission Denied: You are not a signatory to this cocoon")
	}

	// convert emails to identity ids
	for i, id := range req.IDs {
		if govalidator.IsEmail(id) {
			req.IDs[i] = (&types.Identity{Email: id}).GetID()
		}
	}

	var newSignatories []string
	for _, id := range cocoon.Signatories {
		if !util.InStringSlice(req.IDs, id) {
			newSignatories = append(newSignatories, id)
		}
	}

	cocoon.Signatories = newSignatories
	err = api.platform.PutCocoon(ctx, cocoon)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   cocoon.ToJSON(),
	}, nil
}
