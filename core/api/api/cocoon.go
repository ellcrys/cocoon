package api

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/fatih/structs"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
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
func (api *API) CreateCocoon(ctx context.Context, req *proto_api.CocoonPayloadRequest) (*proto_api.Response, error) {

	var err error
	var claims jwt.MapClaims
	var releaseID = util.UUID4()
	var now = time.Now()

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	loggedInIdentity := claims["identity"].(string)
	var cocoon = types.Cocoon{
		ID:             req.ID,
		Memory:         int(req.Memory),
		CPUShare:       int(req.CPUShare),
		SigThreshold:   int(req.SigThreshold),
		NumSignatories: int(req.NumSignatories),
	}

	cocoon.Status = CocoonStatusCreated
	cocoon.Releases = []string{releaseID}
	cocoon.IdentityID = loggedInIdentity
	cocoon.Signatories = append(cocoon.Signatories, cocoon.IdentityID)
	cocoon.CreatedAt = now.UTC().Format(time.RFC3339Nano)

	// ensure a similar cocoon does not exist
	_, err = api.platform.GetCocoon(ctx, cocoon.ID)
	if err != nil && err != types.ErrCocoonNotFound {
		return nil, err
	} else if err == nil {
		return nil, fmt.Errorf("cocoon with matching ID already exists")
	}

	if err := ValidateCocoon(&cocoon); err != nil {
		return nil, err
	}

	err = api.platform.PutCocoon(ctx, &cocoon)
	if err != nil {
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

	// create new release
	var release = &types.Release{
		ID:         releaseID,
		CocoonID:   cocoon.ID,
		URL:        req.URL,
		Version:    req.Version,
		Language:   req.Language,
		BuildParam: req.BuildParam,
		Link:       req.Link,
		Env:        req.Env,
		CreatedAt:  now.UTC().Format(time.RFC3339Nano),
	}

	// Resolve firewall rules destination
	if len(req.Firewall) > 0 {
		var firewall types.Firewall
		for _, f := range req.Firewall {
			firewall = append(firewall, types.FirewallRule{Destination: f.Destination, DestinationPort: f.DestinationPort, Protocol: f.Protocol})
		}
		resolvedFirewall, err := common.ResolveFirewall(firewall)
		if err != nil {
			return nil, fmt.Errorf("Firewall: %s", err)
		}
		release.Firewall = resolvedFirewall
	}

	// set acl map if provided
	if len(req.ACL) > 0 {
		release.ACL, err = types.NewACLMapFromByte(req.ACL)
		if err != nil {
			return nil, fmt.Errorf("acl: %s", err)
		}
	}

	// set firewall if provided
	if len(req.Firewall) > 0 {
		for _, f := range req.Firewall {
			release.Firewall = append(release.Firewall, types.FirewallRule{
				Destination:     f.Destination,
				DestinationPort: f.DestinationPort,
				Protocol:        f.Protocol,
			})
		}
	}

	err = api.platform.PutRelease(ctx, release)
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

	return &proto_api.Response{
		Status: 200,
		Body:   cocoon.ToJSON(),
	}, nil
}

// UpdateCocoon updates a cocoon and optionally creates a new
// release. A new release is created when Release related fields are
// changed.
func (api *API) UpdateCocoon(ctx context.Context, req *proto_api.CocoonPayloadRequest) (*proto_api.Response, error) {

	var err error
	var claims jwt.MapClaims
	var cocoonUpdated bool

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	cocoon, err := api.platform.GetCocoon(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Ensure the cocoon identity matches the logged in user
	loggedInIdentity := claims["identity"].(string)
	if cocoon.IdentityID != loggedInIdentity {
		return nil, fmt.Errorf("Permission denied: You do not have permission to perform this operation")
	}

	var cocoonUpd = types.Cocoon{
		ID:             req.ID,
		Memory:         int(req.Memory),
		CPUShare:       int(req.CPUShare),
		SigThreshold:   int(req.SigThreshold),
		NumSignatories: int(req.NumSignatories),
	}

	// check if the updates defer from the existing cocoon fields values.
	// validate cocoon changes
	if cocoon.HasFieldsChanged(structs.New(cocoonUpd).Map()) {
		cocoonUpdated = true
		cocoon.Memory = cocoonUpd.Memory
		cocoon.CPUShare = cocoonUpd.CPUShare
		cocoon.SigThreshold = cocoonUpd.SigThreshold
		cocoon.NumSignatories = cocoonUpd.NumSignatories

		if err = ValidateCocoon(cocoon); err != nil {
			return nil, err
		}
	}

	// get the last deployed release. if no recently deployed release, get the most recent release
	var recentReleaseID = cocoon.LastDeployedReleaseID
	if len(recentReleaseID) == 0 {
		recentReleaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}

	release, err := api.platform.GetRelease(ctx, recentReleaseID, false)
	if err != nil {
		return nil, err
	}

	var releaseUpdated bool
	releaseUpd := types.Release{
		URL:        req.URL,
		Version:    req.Version,
		Language:   req.Language,
		BuildParam: req.BuildParam,
		Link:       req.Link,
		Env:        req.Env,
	}

	// If ACL is set, create an ACLMap, set the cocoonUpd.ACL
	if len(req.ACL) > 0 {
		releaseUpd.ACL, err = types.NewACLMapFromByte(req.ACL)
		if err != nil {
			return nil, fmt.Errorf("acl: %s", err)
		}
	}

	if len(req.Firewall) > 0 {
		for _, f := range req.Firewall {
			releaseUpd.Firewall = append(releaseUpd.Firewall, types.FirewallRule{Destination: f.Destination, DestinationPort: f.DestinationPort, Protocol: f.Protocol})
		}
		outputFirewall, err := common.ResolveFirewall(releaseUpd.Firewall.DeDup())
		if err != nil {
			return nil, fmt.Errorf("Firewall: %s", err)
		}
		releaseUpd.Firewall = outputFirewall
	}

	// check if the updates defer from the existing last deployed or created fields values.
	// validate cocoon changes
	if release.HasFieldsChanged(structs.New(releaseUpd).Map()) {
		releaseUpdated = true
		release.URL = releaseUpd.URL
		release.Version = releaseUpd.Version
		release.Language = releaseUpd.Language
		release.BuildParam = releaseUpd.BuildParam
		release.Link = releaseUpd.Link
		release.Env = releaseUpd.Env
		release.ACL = releaseUpd.ACL
		release.Firewall = releaseUpd.Firewall

		if err = ValidateCocoon(cocoon); err != nil {
			return nil, err
		}
	}

	var finalResp = map[string]interface{}{
		"newReleaseID":  "",
		"cocoonUpdated": cocoonUpdated,
	}

	// create new release if a field was changed
	if releaseUpdated {

		// reset
		release.ID = util.UUID4()
		release.VotersID = []string{}
		release.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

		// add id to cocoon's releases
		cocoon.Releases = append(cocoon.Releases, release.ID)

		// validate release
		if err = ValidateRelease(release); err != nil {
			return nil, err
		}

		// persist release
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

	apiLog.Infof("Received request to stop cocoon = %s", req.ID)

	var claims jwt.MapClaims
	var err error

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	cocoon, err := api.platform.GetCocoon(ctx, req.GetID())
	if err != nil {
		return nil, err
	}

	// ensure session identity matches cocoon identity
	if claims["identity"].(string) != cocoon.IdentityID {
		return nil, fmt.Errorf("Permission denied: You do not have permission to perform this operation")
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
	var claims jwt.MapClaims
	var err error

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	cocoon, err := api.platform.GetCocoon(ctx, req.CocoonID)
	if err != nil {
		return nil, err
	}

	// ensure session identity matches cocoon identity
	if claims["identity"].(string) != cocoon.IdentityID {
		return nil, fmt.Errorf("Permission denied: You do not have permission to perform this operation")
	}

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

	var claims jwt.MapClaims
	var err error

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

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

	loggedInUserIdentity := claims["identity"].(string)

	// ensure logged in user is a signatory of this cocoon
	if !util.InStringSlice(cocoon.Signatories, loggedInUserIdentity) {
		return nil, fmt.Errorf("Permission Denied: You are not a signatory to this cocoon")
	}

	// ensure logged in user has not voted before
	if release.VotersID != nil && util.InStringSlice(release.VotersID, loggedInUserIdentity) {
		return nil, fmt.Errorf("You have already cast a vote for this release")
	}

	if req.Vote == "1" {
		release.SigApproved++
	}
	if req.Vote == "0" {
		release.SigDenied++
	}

	if release.VotersID == nil {
		release.VotersID = []string{loggedInUserIdentity}
	} else {
		release.VotersID = append(release.VotersID, loggedInUserIdentity)
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

	var claims jwt.MapClaims
	var err error

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	cocoon, err := api.platform.GetCocoon(ctx, req.CocoonID)
	if err != nil {
		return nil, err
	}

	loggedInUserIdentity := claims["identity"].(string)

	// ensure logged user is owner
	if loggedInUserIdentity != cocoon.IdentityID {
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

// FirewallAllow an 'allow' firewall rule to a cocoon
func (api *API) FirewallAllow(ctx context.Context, req *proto_api.FirewallAllowRequest) (*proto_api.Response, error) {
	return nil, fmt.Errorf("not implemented")
}
