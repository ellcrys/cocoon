package api

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/ellcrys/cocoon/core/api/api/proto_api"
	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/types"
	"golang.org/x/crypto/bcrypt"
	context "golang.org/x/net/context"
)

// makeAuthToken creates a session token
func makeAuthToken(id, identity, _type string, exp int64, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"id":       id,
		"identity": identity,
		"type":     _type,
		"exp":      exp,
	})
	return token.SignedString([]byte(secret))
}

// Login authenticates a user and returns a JWT token
func (api *API) Login(ctx context.Context, req *proto_api.LoginRequest) (*proto_api.Response, error) {

	identity, err := api.platform.GetIdentity(ctx, types.NewIdentity(req.GetEmail(), "").GetID())
	if err != nil {
		if common.CompareErr(err, types.ErrIdentityNotFound) == 0 {
			return nil, fmt.Errorf("email or password are invalid")
		}
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(req.GetPassword())); err != nil {
		return nil, fmt.Errorf("email or password are invalid")
	}

	sessionID := util.Sha256(util.UUID4())
	identity.ClientSessions = append(identity.ClientSessions, sessionID)
	key := util.Env("API_SIGN_KEY", "")
	ss, err := makeAuthToken(sessionID, identity.GetID(), "token.cli", time.Now().AddDate(0, 1, 0).Unix(), key)
	if err != nil {
		apiLog.Error(err.Error())
		return nil, fmt.Errorf("failed to create session")
	}

	err = api.platform.PutIdentity(ctx, identity)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Status: 200,
		Body:   []byte(ss),
	}, nil
}

// DeleteSessions deletes all identity sessions or the sessions specified in IDs
func (api *API) DeleteSessions(ctx context.Context, req *proto_api.DeleteSessionsRequest) (*proto_api.Response, error) {

	var err error
	var loggedInIdentity = ctx.Value(types.CtxIdentity).(string)
	var currentSessionToken = ctx.Value(types.CtxSessionID).(string)

	identity, err := api.platform.GetIdentity(ctx, loggedInIdentity)
	if err != nil {
		return nil, err
	}

	if !req.All {
		newSessions := []string{}
		for _, sID := range identity.ClientSessions {
			if sID != currentSessionToken {
				newSessions = append(newSessions, sID)
			}
		}
		identity.ClientSessions = newSessions
	} else {
		identity.ClientSessions = nil
	}

	err = api.platform.PutIdentity(ctx, identity)
	if err != nil {
		return nil, fmt.Errorf("failed to put identity: %s", err)
	}

	return &proto_api.Response{
		Status: 200,
		Body:   []byte("success"),
	}, nil
}
