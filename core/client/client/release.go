package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"strings"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
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

	var releaseID = id
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	ctx = metadata.NewContext(ctx, metadata.Pairs("access_token", userSession.Token))

	// if id is a cocoon id, get the cocoon's most recent release
	if isCocoonID {

		resp, err := client.GetCocoon(ctx, &proto.GetCocoonRequest{ID: id})
		if err != nil {
			stopSpinner()
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				return fmt.Errorf("%s: cocoon does not exists", common.GetShortID(id))
			}
			return err
		}

		util.FromJSON(resp.Body, &cocoon)

		// set id to latest release
		releaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}

	stopSpinner()
	time.Sleep(100 * time.Millisecond)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Are you sure you? Y/n: ")
		text, _ := reader.ReadString('\n')
		v := strings.TrimSpace(strings.ToLower(text))
		if v == "n" {
			log.Info("Aborted")
			return nil
		}
		if v == "y" {
			break
		}
	}

	stopSpinner = util.Spinner("Please wait")

	_, err = client.AddVote(ctx, &proto.AddVoteRequest{
		ReleaseID: releaseID,
		Vote:      vote,
	})
	if err != nil {
		stopSpinner()
		if common.CompareErr(err, fmt.Errorf("release not found")) == 0 {
			return fmt.Errorf("release (%s) was not found", common.GetShortID(releaseID))
		}
		return err
	}

	stopSpinner()

	voteTxt := ""
	switch vote {
	case "1":
		voteTxt = "Approve"
	case "0":
		voteTxt = "Deny"
	}

	log.Info("==> You have successfully voted")
	log.Infof("==> Your vote: %s (%s)", vote, voteTxt)

	return nil
}

// GetReleases fetches one or more releases and logs them
func GetReleases(ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	var releases []types.Release
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
		resp, err = cl.GetRelease(ctx, &proto.GetReleaseRequest{
			ID: id,
		})
		if err != nil {
			if common.CompareErr(err, types.ErrTxNotFound) == 0 {
				stopSpinner()
				err = fmt.Errorf("No such object: %s", id)
				break
			}
			stopSpinner()
			break
		}

		var release types.Release
		if err = util.FromJSON(resp.Body, &release); err != nil {
			return common.JSONCoerceErr("cocoon", err)
		}

		releases = append(releases, release)
		stopSpinner()
	}

	if len(releases) > 0 {
		bs, _ := json.MarshalIndent(releases, "", "   ")
		log.Infof("%s", bs)
	}
	if err != nil {
		return err
	}

	return nil
}
