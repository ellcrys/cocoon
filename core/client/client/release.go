package client

import (
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AddVote adds a new vote to a release. If isCocoonID is true,
// the id is taken to be a cocoon id and as such the vote is added
// to the latest release. A positive vote is denoted with 1 or 0 for
// negative.
func AddVote(id, vote string, isCocoonID bool) error {

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	var release types.Release
	var cocoon types.Cocoon

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()
	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		stopSpinner()
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	client := proto.NewAPIClient(conn)
	md := metadata.Pairs("access_token", userSession.Token)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = metadata.NewContext(ctx, md)

	// if id is a cocoon id, get the cocoon's most recent release
	if isCocoonID {

		resp, err := client.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: id,
		})
		if err != nil {
			stopSpinner()
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				return fmt.Errorf("the cocoon (%s) was not found", common.GetShortID(id))
			}
			return err
		}
		if err = util.FromJSON(resp.Body, &cocoon); err != nil {
			stopSpinner()
			return common.JSONCoerceErr("cocoon", err)
		}

		// get the latest release
		resp, err = client.GetRelease(ctx, &proto.GetReleaseRequest{ReleaseID: cocoon.Releases[len(cocoon.Releases)-1]})
		if err != nil {
			stopSpinner()
			if common.CompareErr(err, types.ErrTxNotFound) == 0 {
				return fmt.Errorf("the release (%s) was not found", common.GetShortID(id))
			}
			return err
		}
		if err = util.FromJSON(resp.Body, &release); err != nil {
			stopSpinner()
			return common.JSONCoerceErr("release", err)
		}
	}

	// fetch release if not already fetched
	if release.ID == "" {
		resp, err := client.GetRelease(ctx, &proto.GetReleaseRequest{ReleaseID: id})
		if err != nil {
			stopSpinner()
			if common.CompareErr(err, types.ErrTxNotFound) == 0 {
				return fmt.Errorf("the release (%s) was not found", common.GetShortID(id))
			}
			return err
		}

		if err = util.FromJSON(resp.Body, &release); err != nil {
			stopSpinner()
			return common.JSONCoerceErr("release", err)
		}
	}

	// get the cocoon
	resp, err := client.GetCocoon(ctx, &proto.GetCocoonRequest{
		ID: release.CocoonID,
	})
	if err != nil {
		stopSpinner()
		if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
			return fmt.Errorf("the cocoon (%s) associated with the release (%s) was not found", common.GetShortID(release.CocoonID), common.GetShortID(id))
		}
		return err
	}
	if err = util.FromJSON(resp.Body, &cocoon); err != nil {
		stopSpinner()
		return common.JSONCoerceErr("cocoon", err)
	}

	// ensure logged in user is a signatory of this cocoon
	loggedInUserIdentity := (&types.Identity{Email: userSession.Email}).GetID()
	if !util.InStringSlice(cocoon.Signatories, loggedInUserIdentity) {
		stopSpinner()
		return fmt.Errorf("Permission Denied: You are not a signatory to this cocoon")
	}

	// ensure logged in user has not voted before
	if release.VotersID != nil && util.InStringSlice(release.VotersID, loggedInUserIdentity) {
		stopSpinner()
		err = fmt.Errorf("You have already sent a vote for this release")
		if isCocoonID {
			err = fmt.Errorf("You have already voted for the latest release of this cocoon")
		}
		return fmt.Errorf("You have already voted for this release")
	}

	if release.VotersID == nil {
		release.VotersID = []string{loggedInUserIdentity}
	} else {
		release.VotersID = append(release.VotersID, loggedInUserIdentity)
	}

	voteTxt := ""
	switch vote {
	case "1":
		release.SigApproved++
		voteTxt = "Approve"
	case "0":
		release.SigDenied++
		voteTxt = "Deny"
	}

	// client := proto.NewAPIClient(conn)
	var protoCreateReleaseReq proto.CreateReleaseRequest
	cstructs.Copy(release, &protoCreateReleaseReq)
	protoCreateReleaseReq.OptionAllowDuplicate = true
	resp, err = client.CreateRelease(ctx, &protoCreateReleaseReq)
	if err != nil {
		stopSpinner()
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	stopSpinner()
	log.Info("==> You have successfully voted")
	log.Infof("==> Your vote: %s (%s)", vote, voteTxt)

	return nil
}
