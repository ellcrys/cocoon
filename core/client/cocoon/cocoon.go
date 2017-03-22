package cocoon

import (
	"fmt"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/client/db"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("api.client")

// APIAddress is the remote address to the cluster server (TODO: change this to production address)
var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8005")

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}

// Create a new cocoon
func Create(cocoon *types.Cocoon) error {

	cocoon.ID = util.UUID4()

	err := api.ValidateCocoon(cocoon)
	if err != nil {
		return err
	}

	userSession, err := db.GetUserSessionToken()
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
	}

	cocoon.Releases = []string{release.ID}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	client := proto.NewAPIClient(conn)
	md := metadata.Pairs("access_token", userSession.Token)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)
	resp, err := client.CreateCocoon(ctx, &proto.CreateCocoonRequest{
		ID:             cocoon.ID,
		URL:            cocoon.URL,
		Language:       cocoon.Language,
		ReleaseTag:     cocoon.ReleaseTag,
		BuildParam:     cocoon.BuildParam,
		Memory:         cocoon.Memory,
		Link:           cocoon.Link,
		CPUShare:       cocoon.CPUShare,
		Releases:       cocoon.Releases,
		NumSignatories: cocoon.NumSignatories,
		SigThreshold:   cocoon.SigThreshold,
	})

	if err != nil {
		stopSpinner()
		if common.CompareErr(err, types.ErrInvalidOrExpiredToken) == 0 {
			return types.ErrClientNoActiveSession
		}
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	resp, err = client.CreateRelease(context.Background(), &proto.CreateReleaseRequest{
		ID:         release.ID,
		CocoonID:   cocoon.ID,
		URL:        cocoon.URL,
		Link:       cocoon.Link,
		Language:   cocoon.Language,
		ReleaseTag: cocoon.ReleaseTag,
		BuildParam: cocoon.BuildParam,
	})

	if err != nil {
		stopSpinner()
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	stopSpinner()
	log.Info("==> New cocoon created")
	log.Infof("==> Cocoon ID: %s", cocoon.ID)
	log.Infof("==> Release ID: %s", release.ID)

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
		ID:         cocoon.ID,
		URL:        cocoon.URL,
		Language:   cocoon.Language,
		ReleaseTag: cocoon.ReleaseTag,
		BuildParam: []byte(cocoon.BuildParam),
		Memory:     cocoon.Memory,
		CPUShare:   cocoon.CPUShare,
	})
	if err != nil {
		return err
	} else if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	return nil
}

// Start starts a new or stopped cocoon code
func Start(id string) error {

	userSession, err := db.GetUserSessionToken()
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
		ID:       id,
		Identity: types.NewIdentity(userSession.Email, "").GetHashedEmail(),
	})

	if err != nil {
		stopSpinner()
		if common.CompareErr(err, types.ErrInvalidOrExpiredToken) == 0 {
			return types.ErrClientNoActiveSession
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
