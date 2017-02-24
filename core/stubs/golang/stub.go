package golang

import (
	"fmt"
	"net"

	"os"

	"github.com/ncodes/cocoon/core/stubs/golang/config"
	"github.com/ncodes/cocoon/core/stubs/golang/proto"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
)

var serverPort = 8000
var defaultServer *stubServer
var log *logging.Logger
var serverDone chan bool

func init() {
	defaultServer = new(stubServer)
	config.ConfigureLogger()
	log = logging.MustGetLogger("ccode.stub")
}

// StartServer starts the stub server and
// readys it for service processing.
// Accepts a callback that is called when the server starts
func StartServer(startedCb func()) {

	serverDone = make(chan bool, 1)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		log.Fatalf("failed to listen on port=%d", serverPort)
	}

	log.Infof("Started stub service at port=%d", serverPort)
	server := grpc.NewServer()
	proto.RegisterStubServer(server, &stubServer{})
	go server.Serve(lis)

	startedCb()
	<-serverDone
	log.Info("Cocoon code stopped")
	os.Exit(0)
}

// Stop stub and cocoon code
func Stop() {
	log.Info("Stopping cocoon code")
	serverDone <- true
}

// StubServer defines the services of the stub's GRPC connection
type stubServer struct {
	port   int
	stream proto.Stub_TransactServer
}

// GetState fetches the value of a blockchain state
func (s *stubServer) Transact(stream proto.Stub_TransactServer) error {
	s.stream = stream
	for {

		in, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("failed to read message from connector. %s", err)
		}

		log.Debug("New message from connector = %s", in.String())

		stream.Send(&proto.Tx{
			Id:   "sample",
			Name: "do something",
		})
	}
}
