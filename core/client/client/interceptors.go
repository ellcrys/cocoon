package client

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/ncodes/cocoon/core/api/api"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ContextTimeout is the duration a context can remain active before it is cancelled
const ContextTimeout = 1 * time.Minute

// Interceptors return the interceptors to run before API calls
func Interceptors() grpc.UnaryClientInterceptor {
	return middleware.ChainUnaryClient(includeAuthorize)
}

// includeAuthorize includes the user session token in the request
func includeAuthorize(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	if util.InStringSlice(api.ExcludeMethodsFromAuth, method) {
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", userSession.Token))
	ctx = metadata.NewOutgoingContext(context.Background(), md)
	return invoker(ctx, method, req, reply, cc, opts...)
}
