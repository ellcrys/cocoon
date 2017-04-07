package api

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	orderer_proto "github.com/ncodes/cocoon/core/orderer/proto"
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

	// CocoonStatusStopped indicates a stopped cocoon
	CocoonStatusStopped = "stopped"
)

// makeCocoonKey constructs a cocoon key
func (api *API) makeCocoonKey(id string) string {
	return fmt.Sprintf("cocoon.%s", id)
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
			return api.putCocoon(ctx, cocoon)
		}
		time.Sleep(500 * time.Millisecond)
	}
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
	odc := orderer_proto.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID:   "",
		LedgerName: types.GetGlobalLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       api.makeCocoonKey(cocoon.ID),
				Value:     string(cocoon.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// CreateCocoon creates a new cocoon and initial release. The new
// cocoon is also added to the identity's list of cocoons
func (api *API) CreateCocoon(ctx context.Context, req *proto.CocoonPayloadRequest) (*proto.Response, error) {

	var err error
	var claims jwt.MapClaims
	var releaseID = util.UUID4()
	var now = time.Now()

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	var cocoon types.Cocoon
	cstructs.Copy(req, &cocoon)
	cocoon.Status = CocoonStatusCreated
	cocoon.Releases = []string{releaseID}
	cocoon.CreatedAt = now.UTC().Format(time.RFC3339Nano)
	req = nil

	cocoon.IdentityID = claims["identity"].(string)
	cocoon.Signatories = append(cocoon.Signatories, cocoon.IdentityID)

	if err := ValidateCocoon(&cocoon); err != nil {
		return nil, err
	}

	// ensure a similar cocoon does not exist
	if _, err = api.GetCocoon(ctx, &proto.GetCocoonRequest{
		ID: cocoon.ID,
	}); err != nil {
		if err != types.ErrCocoonNotFound {
			return nil, err
		} else if err != types.ErrCocoonNotFound {
			return nil, fmt.Errorf("cocoon with matching ID already exists")
		}
	}

	// if a link cocoon id is provided, check if the linked cocoon exists
	// TODO: Provide a permission (ACL) mechanism
	if len(cocoon.Link) > 0 {
		if _, err = api.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: cocoon.Link,
		}); err != nil {
			if err != types.ErrCocoonNotFound {
				return nil, err
			} else if err == types.ErrCocoonNotFound {
				return nil, fmt.Errorf("cannot link to a non-existing cocoon")
			}
		}
	}

	err = api.putCocoon(ctx, &cocoon)
	if err != nil {
		return nil, err
	}

	// create new release
	var protoCreateReleaseReq proto.CreateReleaseRequest
	cstructs.Copy(cocoon, &protoCreateReleaseReq)
	protoCreateReleaseReq.ID = releaseID
	protoCreateReleaseReq.CocoonID = cocoon.ID
	_, err = api.CreateRelease(ctx, &protoCreateReleaseReq)
	if err != nil {
		return nil, err
	}

	// include new cocoon id in identity cocoons list
	identity, err := api.getIdentity(ctx, cocoon.IdentityID)
	if err != nil {
		return nil, err
	}
	identity.Cocoons = append(identity.Cocoons, cocoon.ID)
	err = api.putIdentity(ctx, identity)
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Status: 200,
		Body:   cocoon.ToJSON(),
	}, nil
}

// UpdateCocoon updates a cocoon and optionally creates a new
// release. A new release is created when Release fields are
// set/defined. No release is created if no change was made to previous release.
func (api *API) UpdateCocoon(ctx context.Context, req *proto.CocoonPayloadRequest) (*proto.Response, error) {

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

	// if cocoon has a identity set, ensure the identity matches that of the logged in user
	if len(cocoon.IdentityID) != 0 && cocoon.IdentityID != loggedInIdentity {
		return nil, fmt.Errorf("Permission denied: You do not have permission to perform this operation")
	}

	// update new non-release fields that are different
	// from their existing value
	cocoonUpdated := false
	if len(req.Memory) > 0 && req.Memory != cocoon.Memory {
		cocoon.Memory = req.Memory
		cocoonUpdated = true
	}
	if len(req.CPUShares) > 0 && req.CPUShares != cocoon.CPUShares {
		cocoon.CPUShares = req.CPUShares
		cocoonUpdated = true
	}
	if req.NumSignatories > 0 && req.NumSignatories != cocoon.NumSignatories {
		cocoon.NumSignatories = req.NumSignatories
		cocoonUpdated = true
	}
	if req.SigThreshold > 0 && req.SigThreshold != cocoon.SigThreshold {
		cocoon.SigThreshold = req.SigThreshold
		cocoonUpdated = true
	}

	// validate updated cocoon
	if err = ValidateCocoon(cocoon); err != nil {
		return nil, err
	}

	// get the last deployed release. if no recently deployed release,
	// get the initial release
	var recentReleaseID = cocoon.LastDeployedRelease
	if recentReleaseID == "" {
		recentReleaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}

	resp, err := api.GetRelease(ctx, &proto.GetReleaseRequest{ID: recentReleaseID})
	if err != nil {
		return nil, err
	}

	var release types.Release
	util.FromJSON(resp.Body, &release)

	// Create new release and set values if any of the release field changed
	var releaseUpdated = false
	if len(req.URL) > 0 && req.URL != release.URL {
		release.URL = req.URL
		releaseUpdated = true
	}
	if len(req.ReleaseTag) > 0 && req.ReleaseTag != release.ReleaseTag {
		release.ReleaseTag = req.ReleaseTag
		releaseUpdated = true
	}
	if len(req.Language) > 0 && req.Language != release.Language {
		release.Language = req.Language
		releaseUpdated = true
	}
	if len(req.BuildParam) > 0 && req.BuildParam != release.BuildParam {
		release.BuildParam = req.BuildParam
		releaseUpdated = true
	}
	if len(req.Link) > 0 && req.Link != release.Link {
		release.Link = req.Link
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
		if err = ValidateRelease(&release); err != nil {
			return nil, err
		}

		// persist release
		var protoCreateReleaseReq proto.CreateReleaseRequest
		cstructs.Copy(release, &protoCreateReleaseReq)
		_, err = api.CreateRelease(ctx, &protoCreateReleaseReq)
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

	return &proto.Response{
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

	odc := orderer_proto.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonID: "",
		Key:      api.makeCocoonKey(id),
		Ledger:   types.GetGlobalLedgerName(),
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
func (api *API) GetCocoon(ctx context.Context, req *proto.GetCocoonRequest) (*proto.Response, error) {

	cocoon, err := api.getCocoon(ctx, req.GetID())
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Status: 200,
		Body:   cocoon.ToJSON(),
	}, nil
}

// GetCocoonStatus fetches the cocoon status.It queries the scheduler
// to find out the current service status for the cocoon.
func (api *API) GetCocoonStatus(cocoonID string) (string, error) {

	s, err := api.scheduler.GetServiceDiscoverer().GetByID("cocoon", map[string]string{
		"tag": cocoonID,
	})
	if err != nil {
		apiLog.Errorf("failed to query cocoon service status: %s", err.Error())
		return "", fmt.Errorf("failed to query cocoon service status")
	}

	if len(s) == 0 {
		return CocoonStatusStopped, nil
	}

	return CocoonStatusRunning, nil
}

// StopCocoon stops a running cocoon and sets its status to `stopped`
func (api *API) StopCocoon(ctx context.Context, req *proto.StopCocoonRequest) (*proto.Response, error) {

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

	return &proto.Response{
		Status: 200,
		Body:   []byte("done"),
	}, nil
}

// AddSignatories adds one ore more signatories to a cocoon
func (api *API) AddSignatories(ctx context.Context, req *proto.AddSignatoriesRequest) (*proto.Response, error) {

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
	availableSignatorySlots := cocoon.NumSignatories - int32(len(cocoon.Signatories))
	if availableSignatorySlots < int32(len(req.IDs)) {
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

	return &proto.Response{
		Status: 200,
		Body:   r,
	}, nil
}

// AddVote adds a vote to a release where the logged in user is a signatory
func (api *API) AddVote(ctx context.Context, req *proto.AddVoteRequest) (*proto.Response, error) {

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

	return &proto.Response{
		Status: 200,
		Body:   release.ToJSON(),
	}, nil
}
