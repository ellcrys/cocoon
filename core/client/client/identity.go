package client

import (
	"bytes"
	"fmt"
	"time"

	context "golang.org/x/net/context"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/types"
	"google.golang.org/grpc"
)

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}

// CreateIdentity creates a new identity
func CreateIdentity(email string) error {

	var err error

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}

	stopSpinner := util.Spinner("Please wait")

	client := proto.NewAPIClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Minute)
	resp, err := client.GetIdentity(ctx, &proto.GetIdentityRequest{
		Email: email,
	})

	if err != nil && common.CompareErr(err, types.ErrIdentityNotFound) != 0 {
		stopSpinner()
		return err
	}
	if resp != nil {
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

	log.Info("Please enter your password again")
	password2, err := terminal.ReadPassword(0)
	if err != nil {
		return fmt.Errorf("failed to get password")
	}

	if bytes.Compare(password, password2) != 0 {
		return fmt.Errorf("passwords did not match")
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
	log.Info("==> Email:", email)
	log.Info("==> ID:", (&types.Identity{Email: email}).GetID())

	return nil
}

// AddCocoon adds a cocoon to an identities collection
func AddCocoon(email string, cocoon *types.Cocoon) error {
	return nil
}
