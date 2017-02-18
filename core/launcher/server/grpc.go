package server

import (
	"github.com/ncodes/cocoon/core/launcher/proto"
	logging "github.com/op/go-logging"
	"golang.org/x/net/context"
)

var log = logging.MustGetLogger("connector")

// Server implements the connector grpc interface
type Server struct {
}

// NewServer returns a new connector server
func NewServer() *Server {
	return new(Server)
}

func (srv *Server) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloReply, error) {
	log.Info("Name:", req.GetName())
	return &proto.HelloReply{Message: "How are you?"}, nil
}
