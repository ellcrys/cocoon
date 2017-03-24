package identity

import (
	"fmt"

	context "golang.org/x/net/context"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("api.client")

// APIAddress is the remote address to the cluster server
var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8004")

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}

// Create a new identity
func Create(email string) error {

	var err error

	log.Debug("ADDR: ", APIAddress)
	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}

	stopSpinner := util.Spinner("Please wait")

	client := proto.NewAPIClient(conn)
	resp, err := client.GetIdentity(context.Background(), &proto.GetIdentityRequest{
		Email: email,
	})

	if err != nil && common.ToRPCError(2, types.ErrIdentityNotFound).Error() != err.Error() {
		stopSpinner()
		return err
	} else if resp != nil {
		stopSpinner()
		return types.ErrIdentityAlreadyExists
	}

	stopSpinner()
	log.Info("Enter your password (minimum: 8 characters)")
	password, err := terminal.ReadPassword(0)
	if err != nil {
		return fmt.Errorf("failed to get password")
	}

	if len(password) < 8 {
		stopSpinner()
		return fmt.Errorf("Password is too short. Minimum of 8 characters required")
	}

	stopSpinner = util.Spinner("Please wait")
	resp, err = client.CreateIdentity(context.Background(), &proto.CreateIdentityRequest{
		Email:    email,
		Password: string(password),
	})
	if err != nil {
		stopSpinner()
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	stopSpinner()
	log.Info("==> Successfully created a new identity")
	log.Info("==> ID:", email)

	return nil
}

// AddCocoon adds a cocoon to an identities collection
func AddCocoon(email string, cocoon *types.Cocoon) error {
	return nil
}
