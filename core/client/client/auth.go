package client

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("api.client")

// Login authenticates the client user. It sends the credentials
// to the platform and returns a JWT token for future requests.
func Login(email, password string) error {

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}

	stopSpinner := util.Spinner("Please wait")

	client := proto_api.NewAPIClient(conn)
	resp, err := client.Login(context.Background(), &proto_api.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		stopSpinner()
		return err
	}

	if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	userSession := &types.UserSession{
		Email: email,
		Token: string(resp.Body),
	}

	err = GetDefaultDB().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("auth"))
		if err != nil {
			return err
		}
		if err = b.Put([]byte("auth.user_session"), userSession.ToJSON()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		stopSpinner()
		return err
	}

	stopSpinner()
	log.Info("Login successful")

	return nil
}
