package identity

import (
	"fmt"

	"github.com/ellcrys/crypto/ecdsa"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/grpc/proto"
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

// Identity represents a person or an organization
// in on the platform.
type Identity struct {
}

// NewIdentity creates a new Identity
func NewIdentity() *Identity {
	return new(Identity)
}

// Create a new identity
func (i *Identity) Create(email, pubKey string) error {

	var key *ecdsa.SimpleECDSA

	if len(pubKey) == 0 {
		key = ecdsa.NewSimpleECDSA(ecdsa.CurveP256)
		pubKey = key.GetPubKey()
	} else {
		valid, _ := ecdsa.IsValidPubKey(pubKey)
		if !valid {
			log.Fatal("Public key is invalid. Please use the keygen tool to generate keys")
		}
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}

	client := proto.NewAPIClient(conn)
	resp, err := client.CreateIdentity(context.Background(), &proto.CreateIdentityRequest{
		Id:     util.UUID4(),
		Email:  email,
		PubKey: pubKey,
	})

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	log.Info("==> Successfully created a new identity")
	log.Info("==> ID:", email)
	log.Info("==> TxID:", resp.GetId())
	if key != nil {
		log.Infof("==> Your Private Key: \n%s", key.GetPrivKey())
	}
	log.Infof("==> Your Public Key: \n%s", key.GetPubKey())
	log.Info("\n*** Caution! Please key your private key safe. Do not share it. ***")

	return nil
}
