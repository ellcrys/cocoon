package api

import (
	"time"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	orderer_proto "github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	"golang.org/x/crypto/bcrypt"
	context "golang.org/x/net/context"
)

// putIdentity adds a new identity. If another identity with a matching key
// exists, it is effectively shadowed. The identity password is not saved to the
// systems public ledger but on the system private ledger.
func (api *API) putIdentity(ctx context.Context, identity *types.Identity) error {

	var password = identity.Password

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	identity.Password = ""
	createdAt, _ := time.Parse(time.RFC3339Nano, identity.CreatedAt)
	odc := orderer_proto.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPublicLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeIdentityKey(identity.GetID()),
				Value:     string(identity.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	// save identity password alone in the system private ledger
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPrivateLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeIdentityKey(identity.GetID()),
				Value:     password,
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// CreateIdentity creates a new identity. It returns error if identity
// already exists.
func (api *API) CreateIdentity(ctx context.Context, req *proto.CreateIdentityRequest) (*proto.Response, error) {

	var identity types.Identity
	identity.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	cstructs.Copy(req, &identity)
	req = nil

	if err := ValidateIdentity(&identity); err != nil {
		return nil, err
	}

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	// check if identity already exists
	_, err = api.getIdentity(ctx, identity.GetID())
	if err != nil && err != types.ErrIdentityNotFound {
		return nil, err
	} else if err == nil {
		return nil, types.ErrIdentityAlreadyExists
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(identity.Password), bcrypt.DefaultCost)
	identity.Password = string(hashedPassword)

	err = api.putIdentity(ctx, &identity)
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Status: 200,
		Body:   identity.ToJSON(),
	}, nil
}

// getIdentity gets an existing identity and returns an identity object.
// Since an identity password is never saved along side the rest of the other identity
// field on the system public ledger, it is retrieved from the private system ledger where
// it is stored separately.
func (api *API) getIdentity(ctx context.Context, id string) (*types.Identity, error) {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeIdentityKey(id),
		Ledger:   types.GetSystemPublicLedgerName(),
	})

	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, types.ErrIdentityNotFound
		}
		return nil, err
	}

	passwordTx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeIdentityKey(id),
		Ledger:   types.GetSystemPrivateLedgerName(),
	})
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, fmt.Errorf("identity password transaction not found")
		}
		return nil, err
	}

	var identity types.Identity
	util.FromJSON([]byte(tx.GetValue()), &identity)
	identity.Password = passwordTx.GetValue()

	return &identity, nil
}

// GetIdentity fetches an identity by email or id. If Email field is set in the request,
// it will find the identity by the email (converts the email to an identity key format) or if
// ID field is set, it finds the identity by the id directly.
func (api *API) GetIdentity(ctx context.Context, req *proto.GetIdentityRequest) (*proto.Response, error) {

	var key = req.ID
	if len(req.Email) > 0 {
		key = (&types.Identity{Email: req.Email}).GetID()
	}

	identity, err := api.getIdentity(ctx, key)
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Status: 200,
		Body:   identity.ToJSON(),
	}, nil
}
