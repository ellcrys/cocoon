package api

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// ExcludeMethodsFromAuth includes full method names of
// grpc method to exclude from authentication
var ExcludeMethodsFromAuth = []string{
	"/proto_api.API/CreateIdentity",
	"/proto_api.API/Login",
}

// Interceptors returns the API interceptors
func (api *API) Interceptors() grpc.UnaryServerInterceptor {
	return middleware.ChainUnaryServer(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		apiLog.Debugf("New request [method=%s]", info.FullMethod)
		return handler(ctx, req)
	}, api.authenticateInterceptor)
}

// Authenticate checks whether the request has valid access tokens
func (api *API) authenticateInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	if util.InStringSlice(ExcludeMethodsFromAuth, info.FullMethod) {
		return handler(ctx, req)
	}

	tokenStr, err := common.GetAuthToken(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		tokenType := token.Claims.(jwt.MapClaims)["type"]
		if tokenType == "token.cli" {
			return []byte(util.Env("API_SIGN_KEY", "secret")), nil
		}

		return nil, fmt.Errorf("unknown token type")
	})
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	if err = claims.Valid(); err != nil {
		return nil, err
	}

	identityID := claims["identity"].(string)
	identity, err := api.platform.GetIdentity(ctx, identityID)
	if err != nil {
		if common.CompareErr(err, types.ErrIdentityNotFound) == 0 {
			return nil, fmt.Errorf("invalid session. identity does not exist")
		}
		return nil, err
	}

	sessionID := claims["id"].(string)
	if !util.InStringSlice(identity.ClientSessions, sessionID) {
		return nil, fmt.Errorf("invalid session. Please login")
	}

	ctx = context.WithValue(ctx, types.CtxIdentity, identityID)

	return handler(ctx, req)
}
