package api

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/kr/pretty"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/orderer/proto_orderer"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
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
			return err
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

// putCocoon adds a new cocoon. If another cocoon with a matching key
// exists, it is effectively shadowed
func (api *API) putCocoon(ctx context.Context, cocoon *types.Cocoon) error {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	createdAt, _ := time.Parse(time.RFC3339Nano, cocoon.CreatedAt)
	odc := proto_orderer.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPublicLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeCocoonKey(cocoon.ID),
				Value:     string(cocoon.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

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

	var cocoon types.Cocoon
	cstructs.Copy(req, &cocoon)
	cocoon.Status = CocoonStatusCreated
	cocoon.Releases = []string{releaseID}
	cocoon.IdentityID = loggedInIdentity
	cocoon.Signatories = append(cocoon.Signatories, cocoon.IdentityID)
	cocoon.CreatedAt = now.UTC().Format(time.RFC3339Nano)

	// ensure a similar cocoon does not exist
	_, err = api.GetCocoon(ctx, &proto_api.GetCocoonRequest{ID: cocoon.ID})
	if err != nil && err != types.ErrCocoonNotFound {
		return nil, err
	} else if err == nil {
		return nil, fmt.Errorf("cocoon with matching ID already exists")
	}

	// If ACL is set, create an ACLMap, set the cocoon.ACL
	if len(req.ACL) > 0 {
		var aclMap map[string]interface{}
		if err := util.FromJSON(req.ACL, &aclMap); err != nil {
			return nil, fmt.Errorf("acl: malformed json")
		}
		cocoon.ACL = types.NewACLMap(aclMap)
	}

	if err := ValidateCocoon(&cocoon); err != nil {
		return nil, err
	}

	// Resolve firewall rules destination
	outputFirewall, err := common.ResolveFirewall(cocoon.Firewall)
	if err != nil {
		return nil, fmt.Errorf("Firewall: %s", err)
	}

	cocoon.Firewall = *outputFirewall.DeDup()

	// if a link cocoon id is provided, check if the linked cocoon exists
	// TODO: Provide a permission (ACL) mechanism
	if len(cocoon.Link) > 0 {
		cocoonToLinkTo, err := api.getCocoon(ctx, cocoon.Link)
		if err != nil {
			if err != types.ErrCocoonNotFound {
				return nil, err
			} else if err == types.ErrCocoonNotFound {
				return nil, fmt.Errorf("link: cannot link to a non-existing cocoon %s", cocoon.Link)
			}
		}
		// ensure logged in user owns the cocoon being linked
		if loggedInIdentity != cocoonToLinkTo.IdentityID {
			return nil, fmt.Errorf("link: Permission denied. Cannot create a native link to a cocoon you did not create")
		}
	}

	err = api.putCocoon(ctx, &cocoon)
	if err != nil {
		return nil, err
	}

	// create new release
	var releaseReq proto_api.CreateReleaseRequest
	cstructs.Copy(cocoon, &releaseReq)
	releaseReq.ID = releaseID
	releaseReq.CocoonID = cocoon.ID
	releaseReq.ACL = cocoon.ACL.ToJSON()
	_, err = api.CreateRelease(ctx, &releaseReq)
	if err != nil {
		return nil, err
	}

	// Include new cocoon id in the logged in user's cocoon list
	identity, err := api.getIdentity(ctx, cocoon.IdentityID)
	if err != nil {
		return nil, err
	}
	identity.Cocoons = append(identity.Cocoons, cocoon.ID)
	err = api.putIdentity(ctx, identity)
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

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	cocoon, err := api.getCocoon(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	loggedInIdentity := claims["identity"].(string)

	// Ensure the cocoon identity matches the logged in user
	if cocoon.IdentityID != loggedInIdentity {
		return nil, fmt.Errorf("Permission denied: You do not have permission to perform this operation")
	}

	var cocoonUpd types.Cocoon
	cstructs.Copy(req, &cocoonUpd)

	// If ACL is set, create an ACLMap, set the cocoonUpd.ACL
	if len(req.ACL) > 0 {
		var aclMap map[string]interface{}
		if err := util.FromJSON(req.ACL, &aclMap); err != nil {
			return nil, fmt.Errorf("acl: malformed json")
		}
		cocoonUpd.ACL = types.NewACLMap(aclMap)
	}

	// update new non-release specific fields that have been updated
	cocoonUpdated := false
	if len(cocoonUpd.Memory) > 0 && cocoonUpd.Memory != cocoon.Memory {
		cocoon.Memory = cocoonUpd.Memory
		cocoonUpdated = true
	}
	if len(cocoonUpd.CPUShares) > 0 && cocoonUpd.CPUShares != cocoon.CPUShares {
		cocoon.CPUShares = cocoonUpd.CPUShares
		cocoonUpdated = true
	}
	if cocoonUpd.NumSignatories > 0 && cocoonUpd.NumSignatories != cocoon.NumSignatories {
		cocoon.NumSignatories = cocoonUpd.NumSignatories
		cocoonUpdated = true
	}
	if cocoonUpd.SigThreshold > 0 && cocoonUpd.SigThreshold != cocoon.SigThreshold {
		cocoon.SigThreshold = cocoonUpd.SigThreshold
		cocoonUpdated = true
	}

	if err = ValidateCocoon(cocoon); err != nil {
		return nil, err
	}

	pretty.Println(cocoonUpd.Firewall, "<<<")
	outputFirewall, err := common.ResolveFirewall(*cocoonUpd.Firewall.DeDup())
	if err != nil {
		return nil, fmt.Errorf("Firewall: %s", err)
	}

	cocoonUpd.Firewall = outputFirewall

	// get the last deployed release. if no recently deployed release,
	// get the most recent release
	var recentReleaseID = cocoon.LastDeployedRelease
	if len(recentReleaseID) == 0 {
		recentReleaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}

	release, err := api.getRelease(ctx, recentReleaseID)
	if err != nil {
		return nil, err
	}

	// Create new release and set values if any of the release field changed
	var releaseUpdated = false
	if len(cocoonUpd.URL) > 0 && cocoonUpd.URL != release.URL {
		apiLog.Info("Release changed 1")
		release.URL = cocoonUpd.URL
		releaseUpdated = true
	}
	if len(cocoonUpd.Version) > 0 && cocoonUpd.Version != release.Version {
		release.Version = cocoonUpd.Version
		releaseUpdated = true
	}
	if len(cocoonUpd.Language) > 0 && cocoonUpd.Language != release.Language {
		release.Language = cocoonUpd.Language
		releaseUpdated = true
	}
	if len(cocoonUpd.BuildParam) > 0 && cocoonUpd.BuildParam != release.BuildParam {
		release.BuildParam = cocoonUpd.BuildParam
		releaseUpdated = true
	}
	if len(cocoonUpd.Link) > 0 && cocoonUpd.Link != release.Link {
		release.Link = cocoonUpd.Link
		releaseUpdated = true
	}
	if len(cocoonUpd.Firewall) > 0 && !cocoonUpd.Firewall.Eql(release.Firewall) {
		release.Firewall = cocoonUpd.Firewall
		releaseUpdated = true
	}
	if len(cocoonUpd.ACL) > 0 && !cocoonUpd.ACL.Eql(release.ACL) {
		release.ACL = cocoonUpd.ACL
		releaseUpdated = true
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
		err = api.putRelease(ctx, release)
		if err != nil {
			return nil, err
		}

		finalResp["newReleaseID"] = release.ID
	}

	// update cocoon if cocoon was changed or release was updated/created
	if cocoonUpdated || releaseUpdated {
		err = api.putCocoon(ctx, cocoon)
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

// getCocoon gets a cocoon with a matching id.
func (api *API) getCocoon(ctx context.Context, id string) (*types.Cocoon, error) {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := proto_orderer.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &proto_orderer.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeCocoonKey(id),
		Ledger:   types.GetSystemPublicLedgerName(),
	})
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, types.ErrCocoonNotFound
		}
		return nil, err
	}

	var cocoon types.Cocoon
	util.FromJSON([]byte(tx.Value), &cocoon)

	return &cocoon, nil
}

// GetCocoon fetches a cocoon
func (api *API) GetCocoon(ctx context.Context, req *proto_api.GetCocoonRequest) (*proto_api.Response, error) {

	cocoon, err := api.getCocoon(ctx, req.GetID())
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

	s, err := api.scheduler.GetServiceDiscoverer().GetByID("cocoon", map[string]string{"tag": cocoonID})
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

// StopCocoon stops a running cocoon and sets its status to `stopped`
func (api *API) StopCocoon(ctx context.Context, req *proto_api.StopCocoonRequest) (*proto_api.Response, error) {

	var claims jwt.MapClaims
	var err error

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	cocoon, err := api.getCocoon(ctx, req.GetID())
	if err != nil {
		return nil, err
	}

	// ensure session identity matches cocoon identity
	if claims["identity"].(string) != cocoon.IdentityID {
		return nil, fmt.Errorf("Permission denied: You do not have permission to perform this operation")
	}

	err = api.scheduler.Stop(req.GetID())
	if err != nil {
		apiLog.Error(err.Error())
		return nil, fmt.Errorf("failed to stop cocoon")
	}

	cocoon.Status = CocoonStatusStopped

	if err = api.putCocoon(ctx, cocoon); err != nil {
		apiLog.Error(err.Error())
		return nil, fmt.Errorf("failed to update cocoon status")
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

	cocoon, err := api.getCocoon(ctx, req.CocoonID)
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

		identity, err := api.getIdentity(ctx, id)
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
		err := api.putCocoon(ctx, cocoon)
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

	release, err := api.getRelease(ctx, req.ReleaseID)
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, fmt.Errorf("release not found")
		}
		return nil, err
	}

	cocoon, err := api.getCocoon(ctx, release.CocoonID)
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

	err = api.putRelease(ctx, release)
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

	cocoon, err := api.getCocoon(ctx, req.CocoonID)
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
	err = api.putCocoon(ctx, cocoon)
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
