package golang

import (
	"fmt"
	"net"
	"testing"
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/runtime/golang/proto"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	. "github.com/smartystreets/goconvey/convey"
)

// App is an example cocoon code
type TestCocoonCode struct {
}

// Init method initializes the app
func (c *TestCocoonCode) OnInit(l *Link) error {
	return nil
}

// Invoke process invoke transactions
func (c *TestCocoonCode) OnInvoke(l *Link, txID, function string, params []string) (interface{}, error) {
	return params, nil
}

func startStubServer(t *testing.T, cb func(endCh chan bool, ss *stubServer)) {
	endCh := make(chan bool)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", "7002"))
	if err != nil {
		t.Fatalf("%s", err)
		return
	}
	server := grpc.NewServer()
	stubSrv := new(stubServer)
	proto.RegisterStubServer(server, stubSrv)
	ccode = new(TestCocoonCode)
	go server.Serve(lis)
	cb(endCh, stubSrv)
	<-endCh
	server.Stop()
}

func testClient(t *testing.T, tx *proto.Tx, newReqCb func(*proto.Tx)) {
	conn, err := grpc.Dial(fmt.Sprintf(":%s", "7002"), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client := proto.NewStubClient(conn)
	stream, err := client.Transact(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	defer stream.CloseSend()

	if tx != nil {
		time.AfterFunc(50*time.Millisecond, func() {
			stream.Send(tx)
		})
	}

	for {
		in, _ := stream.Recv()
		newReqCb(in)
		return
	}
}

func TestStub(t *testing.T) {

	startStubServer(t, func(endCh chan bool, ss *stubServer) {
		SetDebugLevel(logging.CRITICAL)
		Convey("GoStub", t, func() {

			Convey(".GetGlobalLedgerName", func() {
				Convey("Should return the expected value set in types.GetGlobalLedgerName()", func() {
					So(GetGlobalLedgerName(), ShouldEqual, types.GetGlobalLedgerName())
				})
			})

			Convey(".isConnected", func() {
				Convey("Should return nil if transaction stream has not been initiated", func() {
					r := isConnected()
					So(r, ShouldEqual, false)
				})
			})

			Convey(".Transact and related methods", func() {

				Convey("Should initialize server stream", func() {
					So(ss.stream, ShouldBeNil)
					go testClient(t, nil, func(tx *proto.Tx) {})
					time.Sleep(50 * time.Millisecond)
					So(ss.stream, ShouldNotBeNil)
				})

				Convey(".sendTx", func() {
					Convey("Should successfully send transaction", func() {
						txToSend := &proto.Tx{Id: util.UUID4(), Invoke: true}
						go func() {
							time.Sleep(50 * time.Millisecond)
							defaultServer = ss
							respChan := make(chan *proto.Tx)
							err := sendTx(txToSend, respChan)
							if err != nil {
								t.Error(".sendTx failed. ", err)
							}
						}()
						testClient(t, nil, func(received *proto.Tx) {
							Convey("Received transaction must have the same values with the one sent", func() {
								So(received, ShouldNotBeNil)
								So(received.Id, ShouldEqual, txToSend.Id)
								So(received.Invoke, ShouldEqual, txToSend.Invoke)

								Convey("Expects the response channel of to be stored", func() {
									So(txRespChannels.Has(txToSend.Id), ShouldEqual, true)
								})
							})
						})
					})
				})

				Convey(".handleInvokeTransaction", func() {

					Convey("Should return error if invoke transaction name is unexpected", func() {
						txToSendFromClient := &proto.Tx{Id: util.UUID4(), Invoke: true, Name: "unexpected"}
						testClient(t, txToSendFromClient, func(received *proto.Tx) {
							So(received.Body, ShouldResemble, []byte("Unsupported invoke transaction named 'unexpected'"))
						})
					})

					Convey("Should fail to call cocoon code invoke function if cocoon code is not running", func() {
						txToSendFromClient := &proto.Tx{
							Id:     util.UUID4(),
							Invoke: true,
							Name:   "function",
							Params: []string{"a", "b", "c"},
						}
						testClient(t, txToSendFromClient, func(received *proto.Tx) {
							So(received, ShouldNotBeNil)
							So(received.Body, ShouldNotBeEmpty)
							So(running, ShouldEqual, false)
							So([]byte("cocoon code is not running"), ShouldResemble, received.Body)
						})
					})

					Convey("Should successfully call cocoon code invoke function", func() {
						running = true
						txToSendFromClient := &proto.Tx{
							Id:     util.UUID4(),
							Invoke: true,
							Name:   "function",
							Params: []string{"a", "b", "c"},
						}
						testClient(t, txToSendFromClient, func(received *proto.Tx) {
							So(received, ShouldNotBeNil)
							So(received.Body, ShouldNotBeEmpty)
							sentParamJSON, _ := util.ToJSON(txToSendFromClient.Params[1:])
							So(sentParamJSON, ShouldResemble, received.Body)
						})
					})
				})

				Convey(".handleRespTransaction", func() {
					Convey("Should return error if transaction has not response channel", func() {
						txToSendFromClient := &proto.Tx{
							Id:       util.UUID4(),
							Response: true,
							Name:     "function",
							Params:   []string{"a", "b", "c"},
						}
						testClient(t, txToSendFromClient, func(received *proto.Tx) {
							So(received, ShouldNotBeNil)
							So(received.Body, ShouldNotBeEmpty)
							So([]byte("response transaction ("+txToSendFromClient.Id+") does not have a corresponding response channel"), ShouldResemble, received.Body)
						})
					})

					Convey("Should successfully receive response in response channel and expects the channel to be remove from the response channel list", func() {
						txToSendFromClient := &proto.Tx{
							Id:       util.UUID4(),
							Response: true,
							Name:     "function",
							Params:   []string{"a", "b", "c"},
						}
						respChan := make(chan *proto.Tx, 1)
						So(txRespChannels.Has(txToSendFromClient.Id), ShouldEqual, false)
						txRespChannels.Set(txToSendFromClient.Id, respChan)
						So(txRespChannels.Has(txToSendFromClient.Id), ShouldEqual, true)
						go testClient(t, txToSendFromClient, func(received *proto.Tx) {})
						tx := <-respChan
						So(tx.Id, ShouldEqual, txToSendFromClient.Id)
						So(txRespChannels.Has(txToSendFromClient.Id), ShouldEqual, false)
					})

				})
			})
		})
		close(endCh)
	})
}
