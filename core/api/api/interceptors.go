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
	"/proto_api.API/GetIdentity",
	"/proto_api.API/Login",
	"/proto_api.API/GetRelease",
}

// Interceptors returns the API interceptors
func Interceptors() grpc.UnaryServerInterceptor {
	return middleware.ChainUnaryServer(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		apiLog.Debugf("New request [method=%s]", info.FullMethod)
		return handler(ctx, req)
	}, authenticate)
}

// Authenticate checks whether the request has valid access tokens
func authenticate(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

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

	ctx = context.WithValue(ctx, types.CtxIdentity, claims["identity"].(string))

	return handler(ctx, req)
}
