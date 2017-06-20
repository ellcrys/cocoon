package client

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/ellcrys/cocoon/core/config"
	"google.golang.org/grpc"
)

// APIAddress is the remote address to the cluster server
var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8004")

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}

// GetAPIConnection returns a connection to the platform API.
// It injects the users session into the connector.
func GetAPIConnection() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(APIAddress, grpc.WithUnaryInterceptor(Interceptors()), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Ellcrys. please try again")
	}
	return conn, nil
}
