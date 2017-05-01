package api

import (
	"time"

	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	"golang.org/x/crypto/bcrypt"
	context "golang.org/x/net/context"
)

// CreateIdentity creates a new identity. It returns error if identity
// already exists.
func (api *API) CreateIdentity(ctx context.Context, req *proto_api.CreateIdentityRequest) (*proto_api.Response, error) {

	var identity types.Identity
	identity.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	cstructs.Copy(req, &identity)
	req = nil

	if err := ValidateIdentity(&identity); err != nil {
		return nil, err
	}

	// check if identity already exists
	_, err := api.platform.GetIdentity(ctx, identity.GetID())
	if err != nil && err != types.ErrIdentityNotFound {
		return nil, err
	} else if err == nil {
		return nil, types.ErrIdentityAlreadyExists
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(identity.Password), bcrypt.DefaultCost)
	identity.Password = string(hashedPassword)

	err = api.platform.PutIdentity(ctx, &identity)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   identity.ToJSON(),
	}, nil
}

// GetIdentity fetches an identity by email or id. If Email field is set in the request,
// it will find the identity by the email (converts the email to an identity key format) or if
// ID field is set, it finds the identity by the id directly.
func (api *API) GetIdentity(ctx context.Context, req *proto_api.GetIdentityRequest) (*proto_api.Response, error) {

	var key = req.ID
	if len(req.Email) > 0 {
		key = (&types.Identity{Email: req.Email}).GetID()
	}

	identity, err := api.platform.GetIdentity(ctx, key)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   identity.ToJSON(),
	}, nil
}
