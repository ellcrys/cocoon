package auth

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/grpc/proto"
	"github.com/ncodes/cocoon/core/client/db"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("api.client")

// APIAddress is the remote address to the cluster server
var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8004")

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}

// Login authenticates the client user. It sends the credentials
// to the platform and returns a JWT token for future requests.
func Login(email, password string) error {

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}

	stopSpinner := util.Spinner("Please wait")

	client := proto.NewAPIClient(conn)
	resp, err := client.Login(context.Background(), &proto.LoginRequest{
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

	err = db.GetDefaultDB().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("auth"))
		if err != nil {
			return err
		}
		if err = b.Put([]byte("user.session_token"), resp.Body); err != nil {
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
