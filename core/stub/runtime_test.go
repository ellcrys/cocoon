package stub

import (
	"fmt"
	"net"
	"testing"

	"github.com/ncodes/cocoon/core/stub/proto_runtime"
	// . "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
)

// App is an example cocoon code
type TestCocoonCode struct {
}

// Init method initializes the app
func (c *TestCocoonCode) OnInit() error {
	return nil
}

// Invoke process invoke transactions
func (c *TestCocoonCode) OnInvoke(m Metadata, function string, params []string) ([]byte, error) {
	return nil, nil
}

func (c *TestCocoonCode) OnStop() {}

func startStubServer(t *testing.T, cb func(endCh chan struct{}, ss *stubServer)) {
	endCh := make(chan struct{})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", "7002"))
	if err != nil {
		t.Fatalf("%s", err)
		return
	}
	server := grpc.NewServer()
	stubSrv := new(stubServer)
	proto_runtime.RegisterStubServer(server, stubSrv)
	ccode = new(TestCocoonCode)
	go server.Serve(lis)
	cb(endCh, stubSrv)
	<-endCh
	server.Stop()
}
