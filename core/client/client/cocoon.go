package client

import (
	"encoding/json"
	"fmt"
	"time"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	"github.com/xeonx/timeago"
)

// createCocoon creates a cocoon. Expects a contex and a connection object.
// If allowDup is set to true, duplicate/existing cocoon key check is ignored and the record
// is overridden
func createCocoon(ctx context.Context, conn *grpc.ClientConn, cocoon *types.Cocoon, allowDup bool) error {

	client := proto.NewAPIClient(conn)
	resp, err := client.CreateCocoon(ctx, &proto.CreateCocoonRequest{
		ID:                   cocoon.ID,
		URL:                  cocoon.URL,
		Language:             cocoon.Language,
		ReleaseTag:           cocoon.ReleaseTag,
		BuildParam:           cocoon.BuildParam,
		Memory:               cocoon.Memory,
		Link:                 cocoon.Link,
		CPUShares:            cocoon.CPUShares,
		Releases:             cocoon.Releases,
		NumSignatories:       cocoon.NumSignatories,
		SigThreshold:         cocoon.SigThreshold,
		Signatories:          cocoon.Signatories,
		CreatedAt:            cocoon.CreatedAt,
		OptionAllowDuplicate: allowDup,
	})

	if err != nil {
		if common.CompareErr(err, types.ErrInvalidOrExpiredToken) == 0 {
			return types.ErrClientNoActiveSession
		}
		return err
	} else if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	return nil
}

// CreateCocoon a new cocoon
func CreateCocoon(cocoon *types.Cocoon) error {

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	err = api.ValidateCocoon(cocoon)
	if err != nil {
		return err
	}

	release := types.Release{
		ID:         util.UUID4(),
		CocoonID:   cocoon.ID,
		URL:        cocoon.URL,
		ReleaseTag: cocoon.ReleaseTag,
		Language:   cocoon.Language,
		BuildParam: cocoon.BuildParam,
		Link:       cocoon.Link,
		VotersID:   []string{},
		CreatedAt:  cocoon.CreatedAt,
	}

	cocoon.Releases = []string{release.ID}

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		stopSpinner()
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	md := metadata.Pairs("access_token", userSession.Token)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)
	if err = createCocoon(ctx, conn, cocoon, false); err != nil {
		stopSpinner()
		return err
	}
	client := proto.NewAPIClient(conn)
	resp, err := client.CreateRelease(ctx, &proto.CreateReleaseRequest{
		ID:         release.ID,
		CocoonID:   cocoon.ID,
		URL:        cocoon.URL,
		Link:       cocoon.Link,
		Language:   cocoon.Language,
		ReleaseTag: cocoon.ReleaseTag,
		BuildParam: cocoon.BuildParam,
		CreatedAt:  cocoon.CreatedAt,
	})

	if err != nil {
		stopSpinner()
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	resp, err = client.GetIdentity(ctx, &proto.GetIdentityRequest{
		Email: userSession.Email,
	})
	if err != nil {
		return err
	}

	var identity types.Identity
	if err = util.FromJSON(resp.Body, &identity); err != nil {
		return common.JSONCoerceErr("identity", err)
	}

	// add cocoon id to the identity and override the identity key
	identity.Cocoons = append(identity.Cocoons, cocoon.ID)
	var protoCreateIdentityReq proto.CreateIdentityRequest
	cstructs.Copy(identity, &protoCreateIdentityReq)
	protoCreateIdentityReq.OptionAllowDuplicate = true
	_, err = client.CreateIdentity(ctx, &protoCreateIdentityReq)
	if err != nil {
		return err
	}

	stopSpinner()
	log.Info(`==> New cocoon created`)
	log.Infof(`==> Cocoon ID:  %s`, cocoon.ID)
	log.Infof(`==> Release ID: %s`, release.ID)

	return nil
}

// GetCocoons fetches one or more cocoons and logs them
func GetCocoons(ids []string) error {

	var cocoons = []types.Cocoon{}
	var err error
	var resp *proto.Response
	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	for _, id := range ids {
		stopSpinner := util.Spinner("Please wait")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cl := proto.NewAPIClient(conn)
		resp, err = cl.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: id,
		})
		if err != nil {
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				stopSpinner()
				err = fmt.Errorf("No such object: %s", id)
				break
			}
			stopSpinner()
			break
		}

		var cocoon types.Cocoon
		if err = util.FromJSON(resp.Body, &cocoon); err != nil {
			return common.JSONCoerceErr("cocoon", err)
		}

		cocoons = append(cocoons, cocoon)
		stopSpinner()
	}

	bs, _ := json.MarshalIndent(cocoons, "", "   ")
	log.Info("%s", bs)
	if err != nil {
		return err
	}

	return nil
}

// Deploy creates and sends a deploy request to the server
func deploy(ctx context.Context, cocoon *types.Cocoon) error {

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	client := proto.NewAPIClient(conn)
	resp, err := client.Deploy(ctx, &proto.DeployRequest{
		CocoonID:   cocoon.ID,
		URL:        cocoon.URL,
		Language:   cocoon.Language,
		ReleaseTag: cocoon.ReleaseTag,
		BuildParam: []byte(cocoon.BuildParam),
		Memory:     cocoon.Memory,
		CPUShares:  cocoon.CPUShares,
		Link:       cocoon.Link,
	})
	if err != nil {
		return err
	} else if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	return nil
}

// ListCocoons fetches and displays running cocoons belonging to
// the logged in user. Set showAll to true to list both running
// and stopped cocoons.
func ListCocoons(showAll, jsonFormatted bool) error {

	var cocoons []types.Cocoon
	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()
	client := proto.NewAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	resp, err := client.GetIdentity(ctx, &proto.GetIdentityRequest{
		Email: userSession.Email,
	})
	if err != nil {
		stopSpinner()
		return err
	}

	var identity types.Identity
	if err = util.FromJSON(resp.Body, &identity); err != nil {
		stopSpinner()
		return common.JSONCoerceErr("identity", err)
	}

	for _, cid := range identity.Cocoons {

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		resp, err = client.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: cid,
		})
		if err != nil {
			stopSpinner()
			return err
		}

		var cocoon types.Cocoon
		if err = util.FromJSON(resp.Body, &cocoon); err != nil {
			stopSpinner()
			return common.JSONCoerceErr("cocoon", err)
		}

		cocoons = append(cocoons, cocoon)
	}

	stopSpinner()

	if jsonFormatted {
		bs, _ := json.MarshalIndent(cocoons, "", "   ")
		log.Info("%s", bs)
		return nil
	}

	log.Info("COCOON ID\t\t\t\tRELEASE ID\t\tCREATED\t\t\tSTATUS")
	for _, cocoon := range cocoons {
		created, _ := time.Parse(time.RFC3339, cocoon.CreatedAt)
		log.Infof("%s\t%s\t\t%s\t\t%s",
			cocoon.ID,
			common.GetShortID(cocoon.Releases[len(cocoon.Releases)-1]),
			common.CapitalizeString(timeago.English.Format(created)),
			"Up 10 minutes",
		)
	}

	return nil
}

// Start starts a new or stopped cocoon code
func Start(id string) error {

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	md := metadata.Pairs("access_token", userSession.Token)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	cl := proto.NewAPIClient(conn)
	resp, err := cl.GetCocoon(ctx, &proto.GetCocoonRequest{
		ID: id,
	})

	if err != nil {
		stopSpinner()
		if common.CompareErr(err, types.ErrInvalidOrExpiredToken) == 0 {
			return types.ErrClientNoActiveSession
		} else if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
			return fmt.Errorf("the cocoon (%s) was not found", common.GetShortID(id))
		}
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	var cocoon types.Cocoon
	err = util.FromJSON(resp.Body, &cocoon)

	if err = deploy(ctx, &cocoon); err != nil {
		stopSpinner()
		return err
	}

	stopSpinner()
	log.Info("==> Successfully created a deployment request")
	log.Info("==> ID:", cocoon.ID)

	return nil
}

// AddSignatories adds one or more valid identites to a cocoon's signatory list.
// All valid identities are included and invalid ones will process an error log..
func AddSignatories(cocoonID string, ids []string) error {

	var validIDs []string

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	md := metadata.Pairs("access_token", userSession.Token)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	cl := proto.NewAPIClient(conn)
	resp, err := cl.GetCocoon(ctx, &proto.GetCocoonRequest{
		ID: cocoonID,
	})
	if err != nil {
		stopSpinner()
		if common.CompareErr(err, types.ErrInvalidOrExpiredToken) == 0 {
			return types.ErrClientNoActiveSession
		} else if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
			return fmt.Errorf("the cocoon (%s) was not found", common.GetShortID(cocoonID))
		}
		return err
	}

	var cocoon types.Cocoon
	if err = util.FromJSON(resp.Body, &cocoon); err != nil {
		stopSpinner()
		return common.JSONCoerceErr("cocoon", err)
	}

	// find identity and included in cccoon signatories field
	for _, id := range ids {

		var req = proto.GetIdentityRequest{ID: id}
		shortID := common.GetShortID(id)
		if govalidator.IsEmail(id) {
			req.Email = id
			req.ID = ""
			id = (&types.Identity{Email: id}).GetID()
			shortID = common.GetShortID(id)
		}

		_, err := cl.GetIdentity(ctx, &req)
		if err != nil {
			stopSpinner()
			if common.CompareErr(err, types.ErrIdentityNotFound) == 0 {
				log.Infof("Warning: Identity (%s) is unknown. Skipped.", shortID)
				continue
			} else {
				return fmt.Errorf("failed to get identity: %s", err)
			}
		}
		if util.InStringSlice(cocoon.Signatories, id) {
			stopSpinner()
			log.Infof("Warning: Identity (%s) is already a signatory. Skipped.", shortID)
			continue
		}

		validIDs = append(validIDs, id)
	}

	// append valid ides to the cocoon's existing signatories
	cocoon.Signatories = append(cocoon.Signatories, validIDs...)

	conn, err = grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		stopSpinner()
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	if err = createCocoon(ctx, conn, &cocoon, true); err != nil {
		stopSpinner()
		return err
	}

	stopSpinner()

	if len(validIDs) == 0 {
		log.Info("No new signatory was added")
	} else if len(validIDs) == 1 {
		log.Info(`==> Successfully added a signatory:`)
	} else {
		log.Info(`==> Successfully added the following signatories:`)
	}

	for i, id := range validIDs {
		log.Infof(`==> %d. %s`, i+1, id)
	}

	return nil
}
